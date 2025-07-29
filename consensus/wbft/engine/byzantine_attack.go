package wbftengine

import (
	"bytes"
	"encoding/hex"
	"fmt"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/consensus/wbft/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
)

func (e *Engine) getByzantineExecutableAttacks() map[btypes.AttackType]*btypes.ExecutableAttack {
	hook := e.backend.ByzantineHook()
	if hook == nil {
		return nil
	}

	c := e.backend.Core()
	if c == nil {
		log.Trace("[BYZ] skipping: core is nil", "coreNil", c == nil)
		return nil
	}

	curView := c.CurrentView()
	if curView == nil {
		log.Trace("[BYZ] skipping: curView is nil", "curViewNil", c == nil)
		return nil
	}

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	if state := c.GetState(); state == core.StateAcceptRequest && c.IsProposer() {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, curView.Sequence.Uint64(), curView.Round.Uint64())
	} else {
		log.Trace("[BYZ] skipping: invalid state or not proposer", "have", state, "want", core.StateAcceptRequest, "isProposer", c.IsProposer())
		return nil
	}
	return attacks
}

func (e *Engine) getByzantineAttack(attackType btypes.AttackType,
	attacks map[btypes.AttackType]*btypes.ExecutableAttack) (*btypes.ExecutableAttack, bool) {
	if attacks == nil {
		return nil, false
	}

	at, exists := attacks[attackType]
	if !exists || at == nil {
		return nil, false
	}

	// Check if attack has valid params based on attack type
	switch attackType {
	case btypes.AttackTypeMessagePolicy:
		return at, at.MessagePolicyParams != nil
	case btypes.AttackTypeOmitMessage:
		return at, at.OmitParams != nil
	case btypes.AttackTypeFakeMessage:
		return at, at.FakeParams != nil
	case btypes.AttackTypeTamperedMessage:
		return at, at.TamperParams != nil
	case btypes.AttackTypeRoleSpoofed:
		return at, at.RoleSpoofParams != nil
	case btypes.AttackTypeReplay:
		return at, at.ReplayParams != nil
	default:
		return nil, false
	}
}

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
		switch fakeField.Target {
		case btypes.TargetPrevPrePareSeal, btypes.TargetPrevCommitSeal:
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
		case btypes.TargetPrePareSeal:
		case btypes.TargetCommitSeal:
		}
	}

	return resultPreparedSeal, resultCommittedSeal, attackCount > 0
}

// processFakeField processes a single fake field and returns updated seals
func (e *Engine) processFakeField(
	preparedSeal, committedSeal *types.WBFTAggregatedSeal,
	fakeField btypes.Field,
	fakeIndexOffset, validatorSize int) (*types.WBFTAggregatedSeal, *types.WBFTAggregatedSeal, bool) {

	switch fakeField.Target {
	case btypes.TargetPrevPrePareSeal, btypes.TargetPrevCommitSeal:
		// Generate fake seal based on value
		fakeSeal, err := e.generateFakeSealFromValue(fakeField.Value, fakeIndexOffset, validatorSize)
		if err != nil {
			log.Debug("[BYZ] Failed to generate fake seal", "err", err, "value", fakeField.Value)
			return preparedSeal, committedSeal, false
		}

		if fakeField.Target == btypes.TargetPrevPrePareSeal && preparedSeal != nil {
			// Only modify prepared seal for PrevPrePareSeal
			return e.addFakeSealToAggregated(preparedSeal, fakeSeal, validatorSize), committedSeal, true
		}

		if fakeField.Target == btypes.TargetPrevCommitSeal && committedSeal != nil {
			// Only modify committed seal for PrevCommitSeal
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

func (e *Engine) applyByzantineAttacksOnEpochBlock(chain consensus.ChainHeaderReader, header *types.Header,
	govState govwbft.StateReader) error {
	// TODO:
	// need to modify
	return e.applyByzantineAttacksWithInvalidEpoch(chain, header, govState)
}

func (e *Engine) applyByzantineAttacksWithInvalidEpoch(chain consensus.ChainHeaderReader, header *types.Header,
	govState govwbft.StateReader) error {
	//log.Info("[BYZ] applyByzantineAttacksWithInvalidEpoch called", "caller", bdebug.GetCallerInfo())
	attacks := e.getByzantineExecutableAttacks()
	if attacks == nil {
		return fmt.Errorf("attacks is nil")
	}
	at, exist := e.getByzantineAttack(btypes.AttackTypeFakeMessage, attacks)
	if !exist {
		return fmt.Errorf("not exist executable attack")
	}

	if at != nil && at.FakeParams != nil {
		for _, fakeField := range at.FakeParams.Fields {
			switch fakeField.Target {
			case btypes.TargetHeaderEpochInfo:
				// Generate fake epoch info based on the value
				fakeEpochInfo, err := e.generateFakeEpochInfo(chain, header, govState, fakeField.Value)
				if err != nil {
					log.Error("[BYZ] Failed to generate fake epoch info", "err", err)
					continue
				}

				// Apply the fake epoch info to header
				_, err = ApplyHeaderWBFTExtra(header, WriteEpochInfo(fakeEpochInfo))
				if err != nil {
					log.Error("[BYZ] Failed to write fake epoch info", "err", err)
					continue
				}

				// Mark attack as executed
				//if err := hook.MarkAttackExecuted(at.UID, curView.Sequence.Uint64()); err != nil {
				//	log.Error("[BYZ] Failed to mark attack executed", "uid", at.UID, "err", err)
				//}

				extra, err := getExtra(header)
				if err == nil && extra != nil && extra.EpochInfo != nil {
					log.Info("[BYZ] attack",
						"name", at.NAME,
						"uid", at.UID,
						"seq", e.backend.Core().CurrentView().Sequence.Uint64(),
						"stakers", len(extra.EpochInfo.Stakers),
						"validators", len(extra.EpochInfo.Validators))
					return nil
				}
			}
		}
	}
	return fmt.Errorf("not exist fake params")
}

// generateFakeEpochInfo generates fake epoch info based on attack configuration
func (e *Engine) generateFakeEpochInfo(chain consensus.ChainHeaderReader, header *types.Header,
	state govwbft.StateReader, value interface{}) (*types.EpochInfo, error) {

	// First, build real epoch info as a base
	realEpochInfo, err := e.buildEpochInfo(chain, header, state)
	if err != nil {
		// If we can't build real epoch info, create a minimal fake one
		return e.createMinimalFakeEpochInfo(), nil
	}

	// If value is nil, return a completely fake epoch info
	if value == nil || value == "nil" {
		return e.createMinimalFakeEpochInfo(), nil
	}

	// Parse value to modify the real epoch info
	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this is a DoS attack configuration
		if dosType, hasDosType := v["dos_type"].(string); hasDosType {
			return e.generateDoSEpochInfo(dosType, v)
		}
		return e.modifyEpochInfo(realEpochInfo, v)
	case string:
		// Check for specific DoS attack types
		if v == "dos_massive" || v == "dos_binary" {
			return e.generateDoSEpochInfo(v, nil)
		}
		return e.createMinimalFakeEpochInfo(), nil
	default:
		// Default: add fake validator
		return e.addFakeValidatorToEpochInfo(realEpochInfo), nil
	}
}

// createMinimalFakeEpochInfo creates a minimal fake epoch info
func (e *Engine) createMinimalFakeEpochInfo() *types.EpochInfo {
	fakeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create minimal fake BLS key (48 bytes)
	fakeBLSKey := make([]byte, 48)
	for i := range fakeBLSKey {
		fakeBLSKey[i] = byte(i)
	}

	return &types.EpochInfo{
		Stakers: []*types.Staker{
			{
				Addr:      fakeAddr,
				Diligence: types.DefaultDiligence,
			},
		},
		Validators:    []uint32{0},
		BLSPublicKeys: [][]byte{fakeBLSKey},
		Stabilizing:   false,
	}
}

// addFakeValidatorToEpochInfo adds a fake validator to existing epoch info
func (e *Engine) addFakeValidatorToEpochInfo(epochInfo *types.EpochInfo) *types.EpochInfo {
	// Create a copy
	newEpochInfo := &types.EpochInfo{
		Stakers:       make([]*types.Staker, len(epochInfo.Stakers)+1),
		Validators:    make([]uint32, len(epochInfo.Validators)+1),
		BLSPublicKeys: make([][]byte, len(epochInfo.BLSPublicKeys)+1),
		Stabilizing:   epochInfo.Stabilizing,
	}

	// Copy existing data
	copy(newEpochInfo.Stakers, epochInfo.Stakers)
	copy(newEpochInfo.Validators, epochInfo.Validators)
	copy(newEpochInfo.BLSPublicKeys, epochInfo.BLSPublicKeys)

	// Add fake validator
	fakeAddr := common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	fakeBLSKey := make([]byte, 48)
	for i := range fakeBLSKey {
		fakeBLSKey[i] = 0xFF
	}

	newEpochInfo.Stakers[len(epochInfo.Stakers)] = &types.Staker{
		Addr:      fakeAddr,
		Diligence: types.DiligenceDenominator * 2, // Max diligence
	}
	newEpochInfo.Validators[len(epochInfo.Validators)] = uint32(len(epochInfo.Stakers))
	newEpochInfo.BLSPublicKeys[len(epochInfo.BLSPublicKeys)] = fakeBLSKey

	return newEpochInfo
}

// modifyEpochInfo modifies epoch info based on configuration
func (e *Engine) modifyEpochInfo(epochInfo *types.EpochInfo, config map[string]interface{}) (*types.EpochInfo, error) {
	// Create a copy
	newEpochInfo := &types.EpochInfo{
		Stakers:       epochInfo.Stakers,
		Validators:    epochInfo.Validators,
		BLSPublicKeys: epochInfo.BLSPublicKeys,
		Stabilizing:   epochInfo.Stabilizing,
	}

	// Modify based on config
	if stab, ok := config["stabilizing"].(bool); ok {
		newEpochInfo.Stabilizing = stab
	}

	if validatorCount, ok := config["validator_count"].(float64); ok {
		// Truncate or extend validators
		count := int(validatorCount)
		if count < len(newEpochInfo.Validators) {
			newEpochInfo.Validators = newEpochInfo.Validators[:count]
			newEpochInfo.BLSPublicKeys = newEpochInfo.BLSPublicKeys[:count]
		}
	}

	return newEpochInfo, nil
}

// generateDoSEpochInfo generates oversized epoch info for DoS attacks
func (e *Engine) generateDoSEpochInfo(dosType string, config map[string]interface{}) (*types.EpochInfo, error) {
	log.Debug("[BYZ] Generating DoS EpochInfo", "type", dosType)

	switch dosType {
	case "dos_massive", "massive_stakers":
		return e.generateMassiveStakersEpoch(config)
	case "dos_binary", "binary_payload":
		return e.generateBinaryPayloadEpoch(config)
	case "huge_bls":
		return e.generateHugeBLSKeysEpoch(config)
	case "mixed":
		return e.generateMixedDoSEpoch(config)
	default:
		// Default massive payload
		return e.generateMassiveStakersEpoch(nil)
	}
}

// generateMassiveStakersEpoch creates epoch info with excessive number of stakers
func (e *Engine) generateMassiveStakersEpoch(config map[string]interface{}) (*types.EpochInfo, error) {
	stakerCount := 50000 // Default: 50k stakers
	if config != nil {
		if count, ok := config["staker_count"].(float64); ok {
			stakerCount = int(count)
		}
	}

	log.Warn("[BYZ] Creating massive stakers epoch", "count", stakerCount)

	stakers := make([]*types.Staker, stakerCount)
	validators := make([]uint32, min(stakerCount, 1000)) // Limit validators
	blsKeys := make([][]byte, len(validators))

	// Generate fake stakers
	for i := 0; i < stakerCount; i++ {
		addr := common.BytesToAddress(randomBytes(20))
		stakers[i] = &types.Staker{
			Addr:      addr,
			Diligence: types.DefaultDiligence + uint64(i%1000),
		}
	}

	// Generate validators and BLS keys
	for i := 0; i < len(validators); i++ {
		validators[i] = uint32(i)
		blsKeys[i] = randomBytes(48) // Normal BLS key size
	}

	return &types.EpochInfo{
		Stakers:       stakers,
		Validators:    validators,
		BLSPublicKeys: blsKeys,
		Stabilizing:   false,
	}, nil
}

// generateBinaryPayloadEpoch embeds binary data in BLS keys
func (e *Engine) generateBinaryPayloadEpoch(config map[string]interface{}) (*types.EpochInfo, error) {
	binarySize := 10 * 1024 * 1024 // Default: 10MB
	if config != nil {
		if size, ok := config["binary_size"].(float64); ok {
			binarySize = int(size)
		}
	}

	log.Debug("[BYZ] Creating binary payload epoch", "size", binarySize)

	// Create fake ELF binary header
	elfHeader := []byte{
		0x7f, 0x45, 0x4c, 0x46, // .ELF magic
		0x02, 0x01, 0x01, 0x00, // 64-bit, little endian
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x3e, 0x00, // x86-64 executable
		0x01, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	// Split binary into chunks
	chunkSize := 100 * 1024 // 100KB per BLS key
	numChunks := (binarySize + chunkSize - 1) / chunkSize

	stakers := make([]*types.Staker, numChunks)
	validators := make([]uint32, numChunks)
	blsKeys := make([][]byte, numChunks)

	// Generate binary chunks
	for i := 0; i < numChunks; i++ {
		size := chunkSize
		if i == numChunks-1 {
			size = binarySize - (i * chunkSize)
		}

		chunk := make([]byte, size)
		if i == 0 {
			// First chunk gets ELF header
			copy(chunk, elfHeader)
			// Fill rest with "code"
			for j := len(elfHeader); j < size; j++ {
				chunk[j] = byte((j * 0x90) % 256) // NOP sled pattern
			}
		} else {
			// Other chunks get random "binary data"
			for j := 0; j < size; j++ {
				chunk[j] = byte((i*j + 0x41) % 256)
			}
		}

		stakers[i] = &types.Staker{
			Addr:      common.BytesToAddress(randomBytes(20)),
			Diligence: types.DefaultDiligence,
		}
		validators[i] = uint32(i)
		blsKeys[i] = chunk
	}

	return &types.EpochInfo{
		Stakers:       stakers,
		Validators:    validators,
		BLSPublicKeys: blsKeys,
		Stabilizing:   false,
	}, nil
}

// generateHugeBLSKeysEpoch creates epoch with oversized BLS keys
func (e *Engine) generateHugeBLSKeysEpoch(config map[string]interface{}) (*types.EpochInfo, error) {
	keySize := 1024 * 1024 // Default: 1MB per key
	keyCount := 100        // Default: 100 keys

	if config != nil {
		if size, ok := config["key_size"].(float64); ok {
			keySize = int(size)
		}
		if count, ok := config["key_count"].(float64); ok {
			keyCount = int(count)
		}
	}

	totalSize := keySize * keyCount
	log.Warn("[BYZ] Creating huge BLS keys epoch",
		"keySize", keySize,
		"keyCount", keyCount,
		"totalSize", totalSize)

	stakers := make([]*types.Staker, keyCount)
	validators := make([]uint32, keyCount)
	blsKeys := make([][]byte, keyCount)

	for i := 0; i < keyCount; i++ {
		stakers[i] = &types.Staker{
			Addr:      common.BytesToAddress(randomBytes(20)),
			Diligence: types.DefaultDiligence,
		}
		validators[i] = uint32(i)
		// Massive BLS key
		blsKeys[i] = randomBytes(keySize)
	}

	return &types.EpochInfo{
		Stakers:       stakers,
		Validators:    validators,
		BLSPublicKeys: blsKeys,
		Stabilizing:   false,
	}, nil
}

// generateMixedDoSEpoch combines multiple DoS techniques
func (e *Engine) generateMixedDoSEpoch(config map[string]interface{}) (*types.EpochInfo, error) {
	// Moderate amounts of everything for cumulative effect
	stakerCount := 10000
	keySize := 50 * 1024 // 50KB per key

	log.Warn("[BYZ] Creating mixed DoS epoch",
		"stakers", stakerCount,
		"keySize", keySize)

	stakers := make([]*types.Staker, stakerCount)
	validators := make([]uint32, min(stakerCount, 500))
	blsKeys := make([][]byte, len(validators))

	// Mix of normal and abnormal data
	for i := 0; i < stakerCount; i++ {
		stakers[i] = &types.Staker{
			Addr:      common.BytesToAddress(randomBytes(20)),
			Diligence: uint64(i) * 1000 % types.DiligenceDenominator,
		}
	}

	for i := 0; i < len(validators); i++ {
		validators[i] = uint32(i * 2) // Skip some indices
		blsKeys[i] = randomBytes(keySize)
	}

	return &types.EpochInfo{
		Stakers:       stakers,
		Validators:    validators,
		BLSPublicKeys: blsKeys,
		Stabilizing:   true,
	}, nil
}

// randomBytes generates random bytes
func randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(i % 256)
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
