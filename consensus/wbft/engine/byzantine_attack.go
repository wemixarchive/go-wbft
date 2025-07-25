package wbftengine

import (
	"bytes"
	"encoding/hex"
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
			at,
		)
		if attackApplied {
			preparedSeal = omittedPreparedSeal
			committedSeal = omittedCommittedSeal
			appliedAttacks = append(appliedAttacks, fmt.Sprintf("%s", at.UID))

			log.Info("[BYZ] byzantine attack triggered",
				"name", at.NAME,
				"uid", at.UID,
				"seq", curView.Sequence.Uint64(),
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
				"params", at.OmitParams,
			)
			hook.MarkAttackExecuted(at.UID, curView.Sequence.Uint64())
		}
	} else {
		log.Trace("[BYZ] Invalid or nil OmitAttackParams")
	}

	// ==== FAKE ATTACK ====
	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		if preparedSeal == nil {
			preparedSeal = mergeSeals(originalPreparedSeal, extraPreparedSeal)
		}
		if committedSeal == nil {
			committedSeal = mergeSeals(originalCommittedSeal, extraCommittedSeal)
		}

		// add fake seal
		fakedPreparedSeal, fakedCommittedSeal, attackApplied := e.applyFakeSealAttackIfExists(
			preparedSeal, committedSeal, at.FakeParams.Fields, validators)

		if attackApplied {
			preparedSeal = fakedPreparedSeal
			committedSeal = fakedCommittedSeal
			appliedAttacks = append(appliedAttacks, fmt.Sprintf("%s", at.UID))

			log.Info("[BYZ] byzantine attack triggered",
				"name", at.NAME,
				"uid", at.UID,
				"seq", curView.Sequence.Uint64(),
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
				"params", at.FakeParams,
			)
			hook.MarkAttackExecuted(at.UID, curView.Sequence.Uint64())
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
	at *btypes.ExecutableAttack) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	var mergedPreparedSeal, mergedCommittedSeal *types.WBFTAggregatedSeal
	var attackExecute bool

	// Only apply if attack is for PrePrepare message or for Propagation message
	switch at.OmitParams.Code {
	case btypes.MessageCodePrepare, btypes.MessageCodeCommit, btypes.MessageCodeRoundChange:
		return nil, nil, false
	case btypes.MessageCodePrePrepare:
		switch at.OmitParams.Cmd {
		case btypes.OmitCommandPrepareSeal:
			mergedPreparedSeal, attackExecute = mergeSealsWithOmitAttack()
			mergedCommittedSeal = mergeSeals(originalCommittedSeal, extraCommittedSeals)
		case btypes.OmitCommandCommitSeal:
			mergedPreparedSeal = mergeSeals(originalPreparedSeal, extraPreparedSeals)
			mergedCommittedSeal, attackExecute = mergeSealsWithOmitAttack()
		case btypes.OmitCommandRoundChange:
		case btypes.OmitCommandPrepareMessage:
		}
	case btypes.MessageCodePropagation:
	default:
		log.Error("[BYZ] omit: unknown omit param code", "code", at.OmitParams.Code)
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
	fakeMessage []btypes.Field, validators wbft.ValidatorSet) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	attackCount := 0
	resultPreparedSeal := preparedSeal
	resultCommittedSeal := committedSeal
	baseValidatorCount := validators.Size()

	// Process each fake field and accumulate results
	for i, fakeField := range fakeMessage {
		// Use incremental indices for each fake seal to avoid conflicts
		fakeIndexOffset := baseValidatorCount + i

		updatedPrepared, updatedCommitted, applied := e.processFakeField(
			resultPreparedSeal,
			resultCommittedSeal,
			fakeField,
			fakeIndexOffset,
			baseValidatorCount,
		)

		if applied {
			resultPreparedSeal = updatedPrepared
			resultCommittedSeal = updatedCommitted
			attackCount++
		}
	}

	return resultPreparedSeal, resultCommittedSeal, attackCount > 0
}

// processFakeField processes a single fake field and returns updated seals
func (e *Engine) processFakeField(
	preparedSeal, committedSeal *types.WBFTAggregatedSeal,
	fakeField btypes.Field,
	fakeIndexOffset, validatorSize int) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	// Generate fake seal based on value
	fakeSeal, err := e.generateFakeSealFromValue(fakeField.Value, fakeIndexOffset, validatorSize)
	if err != nil {
		log.Debug("[BYZ] Failed to generate fake seal", "err", err, "value", fakeField.Value)
		return preparedSeal, committedSeal, false
	}

	switch fakeField.Target {
	case btypes.TargetPrevPrePareSeal:
		// Only modify prepared seal for PrevPrePareSeal
		if preparedSeal != nil {
			return e.addFakeSealToAggregated(preparedSeal, fakeSeal, validatorSize), committedSeal, true
		}

	case btypes.TargetPrevCommitSeal:
		// Only modify committed seal for PrevCommitSeal
		if committedSeal != nil {
			return preparedSeal, e.addFakeSealToAggregated(committedSeal, fakeSeal, validatorSize), true
		}

	case btypes.TargetPrePareSeal:
	case btypes.TargetCommitSeal:
	}

	return preparedSeal, committedSeal, false
}

// generateFakeSealFromValue generates a fake seal based on the provided value
func (e *Engine) generateFakeSealFromValue(value interface{}, fakeIndex, validatorSize int) (wbft.SealData, error) {
	if value == nil || value == "nil" {
		// Generate random fake seal
		return e.createFakeSeal(fakeIndex), nil
	}

	// Try to parse value as a map for specific seal configuration
	switch v := value.(type) {
	case string:
		// If it's a string other than "nil", error
		if v != "nil" {
			sig := value.(string)
			if len(sig) > 2 && sig[:2] == "0x" {
				sig = sig[2:]
			}
			decoded, err := hex.DecodeString(sig)
			if err != nil {
				return wbft.SealData{}, fmt.Errorf("failed to decode signature: %w", err)
			}
			signature := decoded
			sealerIndex := fakeIndex
			if validatorSize > 0 && sealerIndex >= validatorSize {
				sealerIndex = validatorSize - 1
			}
			return wbft.SealData{
				Sealer: uint32(sealerIndex),
				Seal:   signature,
			}, nil
		}
		return e.createFakeSeal(fakeIndex), nil

	default:
		return wbft.SealData{}, fmt.Errorf("unsupported value type: %T", value)
	}
}

// createFakeSeal creates a fake seal with the given sealer index
func (e *Engine) createFakeSeal(fakeIndex int) wbft.SealData {
	// Use the provided fake index which should be outside the valid validator range
	fakeSealerIndex := uint32(fakeIndex)

	return wbft.SealData{
		Sealer: fakeSealerIndex,
		Seal:   e.generateFakeSignature(),
	}
}

// addFakeSealToAggregated adds a fake seal to an aggregated seal
func (e *Engine) addFakeSealToAggregated(seal *types.WBFTAggregatedSeal, fakeSeal wbft.SealData, validatorSize int) *types.WBFTAggregatedSeal {
	// Create new sealers set with fake sealer
	newSealers := make(types.SealerSet, len(seal.Sealers))
	copy(newSealers, seal.Sealers)
	if len(seal.Sealers.GetSealers()) == validatorSize {
		// NOTE:
		// If the number of seals is the same as the number of validators,
		// we can't add a sealer, so we only manipulate signature.
	} else {
		newSealers.SetSealer(fakeSeal.Sealer)
	}

	// append the fake signature
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

func mergeSealsWithOmitAttack() (*types.WBFTAggregatedSeal, bool) {
	emptySeal := &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}}
	return emptySeal, true
}
