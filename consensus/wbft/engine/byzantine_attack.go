package wbftengine

import (
	"bytes"
	"fmt"
	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/consensus/wbft/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
)

func (e *Engine) applyByzantineAttacksToSeals(
	originalPreparedSeal, originalCommittedSeal *types.WBFTAggregatedSeal,
	extraPreparedSeal, extraCommittedSeal []wbft.SealData,
	header *types.Header,
	validators wbft.ValidatorSet) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, []string) {

	appliedAttacks := []string{}

	hook := e.backend.ByzantineHook()
	if hook == nil {
		return nil, nil, appliedAttacks
	}

	c := e.backend.Core()
	if c == nil {
		log.Trace("[BYZ] skipping: core is nil", "coreNil", c == nil)
		return nil, nil, appliedAttacks
	}

	curView := c.CurrentView()
	if curView == nil {
		log.Trace("[BYZ] skipping: curView is nil", "curViewNil", c == nil)
		return nil, nil, appliedAttacks
	}

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	if state := c.GetState(); state == core.StateAcceptRequest && c.IsProposer() {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, curView.Sequence.Uint64(), curView.Round.Uint64())
	} else {
		log.Trace("[BYZ] skipping: invalid state or not proposer", "have", state, "want", core.StateAcceptRequest, "isProposer", c.IsProposer())
		return nil, nil, appliedAttacks
	}

	var preparedSeal, committedSeal *types.WBFTAggregatedSeal

	// ==== OMIT ATTACK ====
	if at := attacks[btypes.AttackTypeOmitMessage]; at != nil && at.OmitParams != nil {
		omittedPreparedSeal, omittedCommittedSeal, attackApplied := e.applyOmitAttackIfExists(
			originalPreparedSeal,
			originalCommittedSeal,
			extraPreparedSeal,
			extraCommittedSeal,
			validators,
			at,
		)
		if attackApplied {
			preparedSeal = omittedPreparedSeal
			committedSeal = omittedCommittedSeal
			appliedAttacks = append(appliedAttacks, fmt.Sprintf("%s", at.UID))

			log.Info("[BYZ] attack omit: broadcast_missing_prevseal_preprepare",
				"name", at.NAME,
				"uid", at.UID,
				"seq", curView.Sequence.Uint64(),
				"round", curView.Round.Uint64(),
				"msgCode", at.OmitParams.Code,
				"cmd", btypes.ParseOmitCommand(at.OmitParams.Code, at.OmitParams.Cmd),
				"options", at.OmitParams.Option,
				"preparedSeal_count", func() int {
					if preparedSeal == nil {
						return 0
					}
					return len(preparedSeal.Sealers.GetSealers())
				}(),
				"committedSeal_count", func() int {
					if committedSeal == nil {
						return 0
					}
					return len(committedSeal.Sealers.GetSealers())
				}(),
			)
		}
	} else {
		log.Trace("[BYZ] Invalid or nil OmitAttackParams")
	}

	// ==== FAKE ATTACK ====
	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		// Check for fake seal attack
		for _, fakeField := range at.FakeParams.FakeMessage {
			if fakeField.FakeTarget == btypes.FakeTargetPrevSeal {
				if preparedSeal == nil {
					preparedSeal = mergeSeals(originalPreparedSeal, extraPreparedSeal)
				}
				if committedSeal == nil {
					committedSeal = mergeSeals(originalCommittedSeal, extraCommittedSeal)
				}
				// add fake seal
				fakedPreparedSeal, fakedCommittedSeal, attackApplied := e.applyFakeSealAttackIfExists(
					preparedSeal, committedSeal, fakeField.Value.(string), validators)

				if attackApplied {
					preparedSeal = fakedPreparedSeal
					committedSeal = fakedCommittedSeal
					appliedAttacks = append(appliedAttacks, fmt.Sprintf("%s", at.UID))

					log.Info("[BYZ] attack fake: broadcast_invalid_prevseal_preprepare",
						"name", at.NAME,
						"uid", at.UID,
						"seq", curView.Sequence.Uint64(),
						"round", curView.Round.Uint64(),
						"msgCode", at.FakeParams.Code,
						"fakeType", fakeField.FakeTarget,
						"preparedSeal_count", func() int {
							if preparedSeal == nil {
								return 0
							}
							return len(preparedSeal.Sealers.GetSealers())
						}(),
						"committedSeal_count", func() int {
							if committedSeal == nil {
								return 0
							}
							return len(committedSeal.Sealers.GetSealers())
						}(),
					)
				}
			}
		}
	} else {
		log.Trace("[BYZ] Invalid or nil FakeAttackParams")
	}

	if len(appliedAttacks) > 0 {
		e.logSealStatus(preparedSeal, committedSeal, validators.QuorumSize(), appliedAttacks)
		for _, uid := range appliedAttacks {
			log.Trace("[BYZ] applyByzantineAttacksToSeals", "uid", uid)
			//err := hook.MarkAttackExecuted(uid, curView.Sequence.Uint64())
			//if err != nil {
			//	log.Error("[BYZ] failed to mark executed", "uid", uid, "err", err)
			//}
		}
	}

	return preparedSeal, committedSeal, appliedAttacks
}

// applyOmitAttackIfExists checks for omit attack and applies it to seals if configured
// Returns (modifiedPreparedSeal, modifiedCommittedSeal, attackExecuted)
func (e *Engine) applyOmitAttackIfExists(
	originalPreparedSeal, originalCommittedSeal *types.WBFTAggregatedSeal,
	extraPreparedSeals, extraCommittedSeals []wbft.SealData,
	validators wbft.ValidatorSet, at *btypes.ExecutableAttack) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	var mergedPreparedSeal, mergedCommittedSeal *types.WBFTAggregatedSeal
	var attackExecute bool

	// Only apply if attack is for PrePrepare message or for Propagation message
	switch at.OmitParams.Code {
	case btypes.MessageCodePrepare, btypes.MessageCodeCommit, btypes.MessageCodeRoundChange:
		return nil, nil, false
	case btypes.MessageCodePrePrepare:
		switch at.OmitParams.Cmd {
		case btypes.OmitCommandPrepareSeal:
			mergedPreparedSeal, attackExecute = mergeSealsWithOmitAttack(
				originalPreparedSeal, extraPreparedSeals, at.OmitParams.Option)
			mergedCommittedSeal = mergeSeals(originalCommittedSeal, extraCommittedSeals)
		case btypes.OmitCommandCommitSeal:
			mergedPreparedSeal = mergeSeals(originalPreparedSeal, extraPreparedSeals)
			mergedCommittedSeal, attackExecute = mergeSealsWithOmitAttack(
				originalCommittedSeal, extraCommittedSeals, at.OmitParams.Option)
		case btypes.OmitCommandRoundChange:
		case btypes.OmitCommandPrepareMessage:
		}
	case btypes.MessageCodePropagation:
	default:
		log.Error("[BYZ] attack omit: unknown omit param code", "code", at.OmitParams.Code)
	}

	return mergedPreparedSeal, mergedCommittedSeal, attackExecute
}

func (e *Engine) getSealerCount(seal *types.WBFTAggregatedSeal) int {
	if seal == nil {
		return 0
	}
	return len(seal.Sealers.GetSealers())
}

func (e *Engine) logSealStatus(preparedSeal, committedSeal *types.WBFTAggregatedSeal,
	quorumSize int, appliedAttacks []string) {

	log.Trace("[BYZ] Byzantine attacks applied to seals",
		"attacks", appliedAttacks,
		"prepared_count", e.getSealerCount(preparedSeal),
		"committed_count", e.getSealerCount(committedSeal),
		"quorum_required", quorumSize,
		"prepared_below_quorum", e.getSealerCount(preparedSeal) < quorumSize,
		"committed_below_quorum", e.getSealerCount(committedSeal) < quorumSize)
}

// applyFakeSealAttack applies fake seal attack by adding non-validator signatures
func (e *Engine) applyFakeSealAttackIfExists(
	preparedSeal, committedSeal *types.WBFTAggregatedSeal,
	value string, validators wbft.ValidatorSet) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	if value == "nil" {
		// Generate a fake sealer index (using validator set size + 1)
		// This ensures we're using an index outside the valid validator range
		fakeSealerIndex := uint32(validators.Size())

		// Generate fake seal data
		fakeSeal := wbft.SealData{
			Sealer: fakeSealerIndex,
			Seal:   e.generateFakeSignature(),
		}

		// Create copies of the original seals and add fake seal
		var fakePreparedSeal, fakeCommittedSeal *types.WBFTAggregatedSeal

		if preparedSeal != nil {
			fakePreparedSeal = e.addFakeSealToAggregated(preparedSeal, fakeSeal)
		}

		if committedSeal != nil {
			fakeCommittedSeal = e.addFakeSealToAggregated(committedSeal, fakeSeal)
		}

		return fakePreparedSeal, fakeCommittedSeal, true
	}

	return preparedSeal, committedSeal, false
}

// addFakeSealToAggregated adds a fake seal to an aggregated seal
func (e *Engine) addFakeSealToAggregated(seal *types.WBFTAggregatedSeal, fakeSeal wbft.SealData) *types.WBFTAggregatedSeal {
	// Create new sealers set with fake sealer
	newSealers := make(types.SealerSet, len(seal.Sealers))
	copy(newSealers, seal.Sealers)
	newSealers.SetSealer(fakeSeal.Sealer)

	// For fake attack, we just append the fake signature
	// In reality, this would create an invalid aggregated signature
	seals := [][]byte{seal.Signature, fakeSeal.Seal}

	// Try to aggregate, but if it fails, just concatenate
	aggregated, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		log.Debug("[BYZ] addFakeSealToAggregated: expected aggregation failure for fake signature", "err", err)
		// For testing purpose, just use the original signature
		return &types.WBFTAggregatedSeal{
			Sealers:   newSealers,
			Signature: seal.Signature, // Keep original signature
		}
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   newSealers,
		Signature: aggregated.Marshal(),
	}
}

// generateFakeSignature generates a valid BLS signature using a fake key
func (e *Engine) generateFakeSignature() []byte {
	// Generate a fake BLS secret key with a fixed seed
	seed := bytes.Repeat([]byte{0x42}, 32) // 32 bytes seed
	fakeSecretKey, err := bls.GenerateKey(seed)
	if err != nil {
		log.Debug("[BYZ] Failed to generate fake BLS key", "err", err)
		// Fallback to simple fake signature
		sig := make([]byte, 96)
		for i := range sig {
			sig[i] = byte(i % 256)
		}
		return sig
	}

	// Create a fake seal message similar to PrepareSeal function
	// This creates a 32-byte hash that looks like a real seal
	fakeBlockHash := bytes.Repeat([]byte{0xAB}, 32)   // Fake block hash
	fakeSealMessage := append(fakeBlockHash, byte(0)) // 0 for SealTypePrepare, 1 for SealTypeCommit

	// Sign with the fake key to get a valid BLS signature
	signature := fakeSecretKey.Sign(fakeSealMessage)

	return signature.Marshal()
}

func mergeSealsWithOmitAttack(
	seal *types.WBFTAggregatedSeal,
	extraSeal []wbft.SealData,
	option uint64) (*types.WBFTAggregatedSeal, bool) {
	switch option {
	case 0:
		emptySeal := &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}}
		return emptySeal, true
	case 1:
		emptySeal := &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}}
		return mergeSeals(emptySeal, extraSeal), true
	case 2:
		return seal, true
	default:
	}
	return nil, false
}
