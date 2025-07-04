package wbftengine

import (
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

	// ==== FAKE ATTACK ====
	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		//
	} else {
		log.Trace("[BYZ] Invalid or nil OmitAttackParams")
	}

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
				"msgCode", btypes.MessageCodePrePrepare,
				"cmd", btypes.ParseOmitCommand(at.OmitParams.Code, at.OmitParams.Cmd),
			)

			log.Trace("[BYZ] Omit attack applied to block seals",
				"block_number", header.Number.Uint64(),
				"prepared_seals", func() int {
					if preparedSeal == nil {
						return 0
					}
					return len(preparedSeal.Sealers)
				}(),
				"committed_seals", func() int {
					if committedSeal == nil {
						return 0
					}
					return len(committedSeal.Sealers)
				}())
		}
	} else {
		log.Trace("[BYZ] Invalid or nil OmitAttackParams")
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

	// Only apply if attack is for PrePrepare message
	if at.OmitParams.Code != btypes.MessageCodePrePrepare {
		return nil, nil, false
	}

	// First, perform normal merge to get the final seals with duplicates removed
	mergedPreparedSeal := mergeSeals(originalPreparedSeal, extraPreparedSeals)
	mergedCommittedSeal := mergeSeals(originalCommittedSeal, extraCommittedSeals)
	attackExecuted := false

	switch at.OmitParams.Cmd {
	case btypes.OmitCommandPrevPrepareSeal:
		// Apply omit attack on prev prepare seals
		if mergedPreparedSeal != nil {
			originalCount := len(mergedPreparedSeal.Sealers)
			mergedPreparedSeal = omitSealsBy(int(at.OmitParams.Cnt), mergedPreparedSeal)

			log.Trace("[BYZ] Applied omit attack on PrevPreparedSeal",
				"attack_uid", at.UID,
				"original_count", originalCount,
				"cnt", at.OmitParams.Cnt,
				"remaining_count", func() int {
					if mergedPreparedSeal == nil {
						return 0
					}
					return len(mergedPreparedSeal.Sealers)
				}())

			// Check if below quorum
			if mergedPreparedSeal != nil && len(mergedPreparedSeal.Sealers) < validators.QuorumSize() {
				attackExecuted = true
			}
		}
	case btypes.OmitCommandPrevCommitSeal:
		// Apply omit attack on prev commit seals
		if mergedCommittedSeal != nil {
			originalCount := len(mergedCommittedSeal.Sealers)
			mergedCommittedSeal = omitSealsBy(int(at.OmitParams.Cnt), mergedCommittedSeal)

			log.Trace("[BYZ] Applied omit attack on PrevCommittedSeal",
				"attack_uid", at.UID,
				"original_count", originalCount,
				"cnt", at.OmitParams.Cnt,
				"remaining_count", func() int {
					if mergedCommittedSeal == nil {
						return 0
					}
					return len(mergedCommittedSeal.Sealers)
				}())

			// Check if below quorum
			if mergedCommittedSeal != nil && len(mergedCommittedSeal.Sealers) < validators.QuorumSize() {
				attackExecuted = true
			}
		}
	case btypes.OmitCommandRoundChange:
	case btypes.OmitCommandPrepareMessage:
	}

	return mergedPreparedSeal, mergedCommittedSeal, attackExecuted
}

func (e *Engine) mergeSealWithOmitAttack(seal *types.WBFTAggregatedSeal, extraSeals []wbft.SealData,
	attacks map[btypes.AttackType]*btypes.ExecutableAttack, sealType string) *types.WBFTAggregatedSeal {

	if at := attacks[btypes.AttackTypeOmitMessage]; at != nil && at.OmitParams != nil {
		// Check if this attack is for PrePrepare message
		if at.OmitParams.Code == btypes.MessageCodePrePrepare {
			// Apply to omit based on cmd and seal type
			if (at.OmitParams.Cmd == 1 && sealType == "prepare") ||
				(at.OmitParams.Cmd == 2 && sealType == "commit") {

				// TODO: remove this log
				log.Debug("[BYZ] Applying omit attack on mergeSeals",
					"attack_uid", at.UID,
					"seal_type", sealType,
					"cmd", at.OmitParams.Cmd,
					"cnt", at.OmitParams.Cnt,
					"original_seal_count", len(seal.Sealers),
					"extra_seal_count", len(extraSeals))

				// If cnt == 0, omit all (don't merge at all)
				if at.OmitParams.Cnt == 0 {
					return nil
				}

				// Omit from extraSeals before merging
				if int(at.OmitParams.Cnt) < len(extraSeals) {
					extraSeals = extraSeals[at.OmitParams.Cnt:]
				} else {
					// If cnt >= extraSeals length, no extra seals to add
					return seal
				}
			}
		}
	}

	// Continue with normal merge
	return mergeSeals(seal, extraSeals)
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

func (e *Engine) getSealerCount(seal *types.WBFTAggregatedSeal) int {
	if seal == nil {
		return 0
	}
	return len(seal.Sealers)
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
func (e *Engine) applyFakeSealAttack(
	preparedSeal, committedSeal *types.WBFTAggregatedSeal,
	params *btypes.FakeAttackParams,
	validators wbft.ValidatorSet) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal) {

	// Add fake sealers
	for i, fakeAddr := range params.FakeSealers {
		fakeSealData := wbft.SealData{
			Sealer: uint32(validators.Size() + i), // Use index beyond validator set
			Seal:   e.generateFakeSignature(),     // Generate fake BLS signature
		}

		if params.SealType == "prepare" || params.SealType == "" {
			// Add fake seal to preparedSeal
			if preparedSeal != nil {
				preparedSeal = e.addFakeSealToAggregated(preparedSeal, fakeSealData)
			}
			log.Info("[BYZ] Added fake seal to PrevPreparedSeal",
				"fake_addr", fakeAddr,
				"sealer_index", fakeSealData.Sealer)
		}
		if params.SealType == "commit" || params.SealType == "" {
			// Add fake seal to committedSeal
			if committedSeal != nil {
				committedSeal = e.addFakeSealToAggregated(committedSeal, fakeSealData)
			}
			log.Info("[BYZ] Added fake seal to PrevCommittedSeal",
				"fake_addr", fakeAddr,
				"sealer_index", fakeSealData.Sealer)
		}
	}

	return preparedSeal, committedSeal
}

// generateFakeSignature generates a fake BLS signature
func (e *Engine) generateFakeSignature() []byte {
	// BLS signature is 96 bytes
	sig := make([]byte, 96)
	// Fill with pseudo-random data
	for i := range sig {
		sig[i] = byte(i % 256)
	}
	return sig
}

func omitSealsBy(cnt int, seal *types.WBFTAggregatedSeal) *types.WBFTAggregatedSeal {
	var seals [][]byte
	sealers := make(types.SealerSet, 0)

	if cnt > 0 {
		sealCount := len(seal.Sealers)
		if cnt < sealCount {
			seals = [][]byte{seal.Signature}
			sealers = make(types.SealerSet, sealCount-cnt)
			copy(sealers[:], seal.Sealers[cnt:])
		}
	} else {
		log.Trace("[BYZ] all seal will be omitted")
	}

	aggregated, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		return seal
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   sealers,
		Signature: aggregated.Marshal(),
	}
}

func omitSeals(seal *types.WBFTAggregatedSeal, extraSeals []wbft.SealData, cnt int, quorumSize int) *types.WBFTAggregatedSeal {
	seals := [][]byte{}
	sealers := make(types.SealerSet, 0)

	if cnt > 0 {
		mergedSeals := mergeSeals(seal, extraSeals)
		if cnt < len(mergedSeals.Sealers) {
			seals = [][]byte{mergedSeals.Signature}
			sealers = make(types.SealerSet, len(mergedSeals.Sealers)-cnt)
			copy(sealers[:], mergedSeals.Sealers[cnt:])
		}
	} else {
		log.Trace("[BYZ] all seal will be omitted")
	}

	aggregated, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		return seal
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   sealers,
		Signature: aggregated.Marshal(),
	}
}

// omitSealsFromAggregated removes cnt sealers from an aggregated seal
func omitSealsFromAggregated(seal *types.WBFTAggregatedSeal, cnt uint64) *types.WBFTAggregatedSeal {
	if seal == nil {
		return nil
	}

	// If cnt == 0, omit all
	if cnt == 0 {
		return nil
	}

	// Get current sealers
	sealerIndices := seal.Sealers.GetSealers()

	// If cnt >= total seals, return nil
	if int(cnt) >= len(sealerIndices) {
		return nil
	}

	// Create new sealer set with remaining sealers
	newSealers := make(types.SealerSet, 0)
	for _, idx := range sealerIndices[cnt:] {
		newSealers.SetSealer(idx)
	}

	// Return new seal with reduced sealers but original signature
	// This will cause validation to fail, which is the intended behavior
	return &types.WBFTAggregatedSeal{
		Sealers:   newSealers,
		Signature: seal.Signature,
	}
}

func printSeal(seal *types.WBFTAggregatedSeal) {
	log.Trace("[BYZ] seal :", "seal", seal.String())
	sealers := seal.Sealers.GetSealers()
	for _, sealer := range sealers {
		log.Trace("[BYZ] seal :", "sealer", sealer)
	}
}
