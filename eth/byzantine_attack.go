package eth

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	wbftBackend "github.com/ethereum/go-ethereum/consensus/wbft/backend"
	wbftcore "github.com/ethereum/go-ethereum/consensus/wbft/core"
)

////// Byzantine Attack For ethHandler

func (h *ethHandler) SetByzantineHook(hook btypes.ConsensusHook) {
	(*handler)(h).byzantineHook = hook
	log.Debug("BYZ: Byzantine hook integrated with eth backend handler")
}

func (h *ethHandler) ByzantineHook() btypes.ConsensusHook {
	if (*handler)(h).byzantineHook == nil {
		log.Warn("BYZ: Byzantine hook is not set, returning nil")
		return nil
	}
	return (*handler)(h).byzantineHook
}

////// Byzantine Attack For Ethereum backend

// SetByzantineHook sets the Byzantine hook
func (s *Ethereum) SetByzantineHook(hook btypes.ConsensusHook) {
	if s.handler != nil {
		s.handler.byzantineHook = hook
		log.Debug("BYZ: Byzantine hook integrated with eth backend handler")
	} else {
		log.Warn("BYZ: handler is nil. can't set byzantine hook")
	}
}

////// Byzantine Attack For handler

// createBlockWithMissingSeals creates a modified block with certain seals removed based on the omit command
func (h *handler) createBlockWithMissingSeals(block *types.Block, cmd uint64) *types.Block {
	// Create a new header with deep copy of all fields
	newHeader := deepCopyHeaderByBlock(block)

	wbftExtra, err := types.ExtractWBFTExtra(newHeader)
	if err != nil {
		log.Warn("BYZ: Failed to extract WBFTExtra", "err", err)
		return nil
	}

	// Modify seals based on command
	// Note: WBFT ignores PreparedSeal and CommittedSeal when calculating block hash
	// So we need to modify other fields to create a different hash
	switch cmd {
	case btypes.OmitCommandPrepareSeal:
		log.Trace("BYZ: Setting PreparedSeal to nil")
		wbftExtra.PreparedSeal = nil
	case btypes.OmitCommandCommitSeal:
		log.Trace("BYZ: Setting CommittedSeal to nil")
		wbftExtra.CommittedSeal = nil
	default:
		log.Warn("BYZ: Unknown omit command", "cmd", cmd)
		return nil
	}

	newExtra, err := rlp.EncodeToBytes(wbftExtra)
	if err != nil {
		log.Warn("BYZ: Failed to encode WBFTExtra", "err", err)
		return nil
	}
	newHeader.Extra = make([]byte, len(newExtra))
	copy(newHeader.Extra, newExtra)

	// Create a new block with the modified header
	return types.NewBlockWithHeader(newHeader).WithBody(block.Transactions(), block.Uncles())
}

func (h *handler) createBlockWithFakeSeals(block *types.Block, fakeParams *btypes.FakeAttackParams) *types.Block {
	// Create a new header with deep copy of all fields
	newHeader := deepCopyHeaderByBlock(block)

	wbftExtra, err := types.ExtractWBFTExtra(newHeader)
	if err != nil {
		log.Warn("BYZ: Failed to extract WBFTExtra", "err", err)
		return nil
	}

	for _, field := range fakeParams.Fields {
		// Parse BLS keys instead of just addresses
		fakeKeys, ok := btypes.ParseFakeBLSKeys(field.Value)
		if !ok {
			log.Warn("BYZ: Invalid fake BLS keys for FakeSeal")
			continue
		}

		switch field.Target {
		case btypes.TargetPrePareSeal:
			fakeSeal := h.generateFakeAggregatedSeal(wbftExtra, fakeKeys, wbftcore.SealTypePrepare, block)
			if fakeSeal != nil {
				wbftExtra.PreparedSeal = fakeSeal
				log.Trace("BYZ: Applied fake prepared seal", "signers", len(fakeKeys))
			}

		case btypes.TargetCommitSeal:
			fakeSeal := h.generateFakeAggregatedSeal(wbftExtra, fakeKeys, wbftcore.SealTypeCommit, block)
			if fakeSeal != nil {
				wbftExtra.CommittedSeal = fakeSeal
				log.Trace("BYZ: Applied fake committed seal", "signers", len(fakeKeys))
			}
		}
	}

	newExtra, err := rlp.EncodeToBytes(wbftExtra)
	if err != nil {
		log.Warn("BYZ: Failed to encode WBFTExtra", "err", err)
		return nil
	}
	newHeader.Extra = make([]byte, len(newExtra))
	copy(newHeader.Extra, newExtra)

	// Create a new block with the modified header
	return types.NewBlockWithHeader(newHeader).WithBody(block.Transactions(), block.Uncles())
}

// generateFakeAggregatedSeal generates a fake aggregated seal by replacing some validator signatures with fake ones
func (h *handler) generateFakeAggregatedSeal(originWBFTExtra *types.WBFTExtra, fakeKeys []*btypes.FakeBLSKey, sealType wbftcore.SealType, block *types.Block) *types.WBFTAggregatedSeal {
	if originWBFTExtra == nil {
		log.Warn("BYZ: OriginWBFTExtra is nil")
		return nil
	}

	// validator list get
	wbftEngine := h.chain.Engine().(*wbftBackend.Backend)
	validators, _ := wbftEngine.Engine().GetValidators(h.chain, block.Header().Number, block.Header().ParentHash, nil)
	validatorsAddressList := validators.AddressList()
	extraPreparedSeal, extraCommittedSeal := wbftEngine.Core().ProcessExtraSeal(block, wbftEngine.Core().PriorRound(), wbftEngine.Core().PriorValidators())

	if len(extraPreparedSeal) == 0 && sealType == wbftcore.SealTypePrepare {
		return nil
	}
	if len(extraCommittedSeal) == 0 && sealType == wbftcore.SealTypeCommit {
		return nil
	}

	// check length validator list
	validatorsCount := len(validatorsAddressList)
	fakeKeysCount := len(fakeKeys)

	var originSealData *[]wbft.SealData
	switch sealType {
	case wbftcore.SealTypePrepare:
		originSealData = &extraPreparedSeal
	case wbftcore.SealTypeCommit:
		originSealData = &extraCommittedSeal
	default:
		originSealData = &extraPreparedSeal
	}

	sortedOriginSealData := make([]wbft.SealData, validatorsCount)
	originSealMap := make(map[uint32]bool)
	for _, sealData := range *originSealData {
		index := sealData.Sealer
		if int(index) < validatorsCount {
			sortedOriginSealData[index] = wbft.SealData{
				Sealer: sealData.Sealer,
				Seal:   append([]byte(nil), sealData.Seal...),
			}
			originSealMap[index] = true
		}
	}

	fakeSealData := make([]wbft.SealData, 0, validatorsCount)
	fakeKeyIndex := 0
	for i := 0; i < validatorsCount; i++ {
		if originSealMap[uint32(i)] {
			// 원본 seal이 있는 경우 그대로 사용
			fakeSealData = append(fakeSealData, sortedOriginSealData[i])
		} else {
			// 빈 슬롯에 fake seal 추가
			var fakeKey *btypes.FakeBLSKey
			if fakeKeyIndex < len(fakeKeys) {
				fakeKey = fakeKeys[fakeKeyIndex]
				fakeKeyIndex++
			} else {
				// fakeKeys가 부족한 경우 랜덤 생성
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				fakeKey = generateFakeBLSKey(address, i)
			}

			fakeSealData = append(fakeSealData, wbft.SealData{
				Sealer: uint32(i),
				Seal:   generateFakeExtraSealSignature(fakeKey, block.Header(), originWBFTExtra.Round, sealType),
			})
		}
	}

	// fakeKeysCount가 validatorsCount보다 크거나 같은 경우 - 모든 seal을 fake로 교체
	if fakeKeysCount >= validatorsCount {
		fakeSealData = make([]wbft.SealData, 0, validatorsCount)
		for i := 0; i < validatorsCount; i++ {
			var fakeKey *btypes.FakeBLSKey
			if i < len(fakeKeys) {
				fakeKey = fakeKeys[i]
			} else {
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				fakeKey = generateFakeBLSKey(address, i)
			}

			fakeSealData = append(fakeSealData, wbft.SealData{
				Sealer: uint32(i),
				Seal:   generateFakeExtraSealSignature(fakeKey, block.Header(), originWBFTExtra.Round, sealType),
			})
		}
	}

	// generate seal prepared or committed
	// return modified seal
	seal := &types.WBFTAggregatedSeal{}
	modifiedSeals := mergeSeals(seal, fakeSealData)

	return modifiedSeals
}

func deepCopyHeaderByBlock(block *types.Block) *types.Header {
	// Get the original header
	header := block.Header()

	// Create a new header with deep copy of all fields
	newHeader := &types.Header{
		ParentHash:  header.ParentHash,
		UncleHash:   header.UncleHash,
		Coinbase:    header.Coinbase,
		Root:        header.Root,
		TxHash:      header.TxHash,
		ReceiptHash: header.ReceiptHash,
		Bloom:       header.Bloom,
		Difficulty:  new(big.Int).Set(header.Difficulty),
		Number:      new(big.Int).Set(header.Number),
		GasLimit:    header.GasLimit,
		GasUsed:     header.GasUsed,
		Time:        header.Time,
		Extra:       make([]byte, len(header.Extra)),
		MixDigest:   header.MixDigest,
		Nonce:       header.Nonce,
	}
	if header.BaseFee != nil {
		newHeader.BaseFee = new(big.Int).Set(header.BaseFee)
	}
	copy(newHeader.Extra, header.Extra)

	return newHeader
}

func mergeSeals(seal *types.WBFTAggregatedSeal, extraSeals []wbft.SealData) *types.WBFTAggregatedSeal {
	if len(extraSeals) == 0 {
		return seal
	}

	seals := [][]byte{}
	if seal.Sealers != nil && len(seal.Signature) > 0 {
		seals = append(seals, seal.Signature)
	}

	sealers := make(types.SealerSet, len(seal.Sealers))
	copy(sealers[:], seal.Sealers[:])

	for _, extraSeal := range extraSeals {
		if extraSeal.Seal == nil || len(extraSeal.Seal) == 0 {
			continue
		}

		if sealers.IsSealer(extraSeal.Sealer) {
			continue
		}
		sealers.SetSealer(extraSeal.Sealer)
		seals = append(seals, extraSeal.Seal)
	}
	if len(seals) == 0 {
		log.Warn("BYZ: No seals to aggregate")
		return seal
	}

	aggregatedSeal, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		log.Warn("BYZ: Failed to aggregate signatures", "err", err)
		return seal
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   sealers,
		Signature: aggregatedSeal.Marshal(),
	}
}

func generateFakeExtraSealSignature(fakeKeys *btypes.FakeBLSKey, header *types.Header, round uint32, sealType wbftcore.SealType) []byte {
	sealMessage := generatePrepareSeal(header, round, sealType)
	signature := fakeKeys.SecretKey.Sign(sealMessage)
	return signature.Marshal()
}

func generateFakeExtraData(fakeKeys []*btypes.FakeBLSKey, validatorCount int, header *types.Header, round uint32, sealType wbftcore.SealType) []wbft.SealData {
	sealData := make([]wbft.SealData, 0)

	for i, key := range fakeKeys {
		if i >= validatorCount {
			break // Don't exceed validator count
		}

		// Generate seal message following protocol
		sealMessage := generatePrepareSeal(header, round, sealType)
		signature := key.SecretKey.Sign(sealMessage)

		sealData = append(sealData, wbft.SealData{
			Sealer: uint32(i), // Use validator index
			Seal:   signature.Marshal(),
		})

		log.Debug("BYZ: Generated fake seal", "index", i, "address", key.Address, "sealType", sealType)
	}

	return sealData
}

// generatePrepareSeal generates the seal message following the protocol
// This is equivalent to PrepareSeal function in consensus/wbft/core/core.go
func generatePrepareSeal(header *types.Header, round uint32, sealType wbftcore.SealType) []byte {
	h := types.CopyHeader(header)
	roundHeader := h.WBFTHashWithRoundNumber(round).Bytes()
	return crypto.Keccak256Hash(append(roundHeader, byte(sealType))).Bytes()
}

// generateFakeBLSKey generates a deterministic BLS key for a validator address
func generateFakeBLSKey(addr common.Address, index int) *btypes.FakeBLSKey {
	// Generate a deterministic seed based on address and index
	seed := make([]byte, 32)
	copy(seed[:20], addr.Bytes())
	binary.BigEndian.PutUint32(seed[20:24], uint32(index))

	secretKey, err := bls.GenerateKey(seed)
	if err != nil {
		log.Warn("BYZ: Failed to generate BLS key for validator", "addr", addr, "err", err)
		// Fallback to random key
		secretKey, _ = bls.GenerateKey(nil)
	}

	return &btypes.FakeBLSKey{
		Address:   addr,
		SecretKey: secretKey,
		PublicKey: secretKey.PublicKey(),
	}
}

// ByzantineAttackResult holds the results of Byzantine attack processing
type ByzantineAttackResult struct {
	FilteredPeers []*ethPeer
	ModifiedBlock *types.Block
	ShouldDrop    bool
}

// GetByzantineHook extracts the Byzantine hook from the consensus engine
func GetByzantineHook(engine consensus.Engine) btypes.ConsensusHook {
	if wbftBackend, ok := engine.(interface{ ByzantineHook() btypes.ConsensusHook }); ok {
		return wbftBackend.ByzantineHook()
	}
	return nil
}

// ProcessByzantineAttacks is the main entry point for Byzantine attack processing
func ProcessByzantineAttacks(h *handler, block *types.Block) *ByzantineAttackResult {
	result := &ByzantineAttackResult{
		FilteredPeers: nil,
		ModifiedBlock: nil,
		ShouldDrop:    false,
	}

	// Get Byzantine hook from consensus engine
	hook := GetByzantineHook(h.engine)
	if hook == nil {
		return result
	}

	blockNum := block.NumberU64()
	attacks := hook.GetExecutableAttacks(btypes.MessageCodePropagation, blockNum, 0)

	// Process message policy attack
	if IsAttackEnabled(attacks, btypes.AttackTypeMessagePolicy) {
		attack := attacks[btypes.AttackTypeMessagePolicy]
		peers := h.peers.peersWithoutBlock(block.Hash())
		filteredPeers, shouldDrop := processMessagePolicyAttack(attack, block, peers, hook)

		result.FilteredPeers = filteredPeers
		result.ShouldDrop = shouldDrop

		if shouldDrop {
			return result // Early return if message should be dropped
		}
	}

	// Process omit attack
	if IsAttackEnabled(attacks, btypes.AttackTypeOmitMessage) {
		attack := attacks[btypes.AttackTypeOmitMessage]
		modifiedBlock := processOmitAttack(attack, block, h, hook)
		if modifiedBlock != nil {
			result.ModifiedBlock = modifiedBlock
		}
	}

	// Process fake attack
	if IsAttackEnabled(attacks, btypes.AttackTypeFakeMessage) {
		attack := attacks[btypes.AttackTypeFakeMessage]
		modifiedBlock := processFakeAttack(attack, block, h, hook)
		if modifiedBlock != nil {
			result.ModifiedBlock = modifiedBlock
		}
	}

	return result
}

// processMessagePolicyAttack handles message policy attacks
func processMessagePolicyAttack(
	attack *btypes.ExecutableAttack,
	block *types.Block,
	peers []*ethPeer,
	hook btypes.ConsensusHook) (filteredPeers []*ethPeer, shouldDrop bool) {

	filteredPeers = nil
	shouldDrop = false
	blockNum := block.NumberU64()
	params := attack.MessagePolicyParams

	for _, field := range params.Fields {
		switch field.Target {
		case btypes.TargetMsgPolicyDirection:
			shouldDrop = checkMessageDirection(field)

		case btypes.TargetMsgPolicyTargets:
			filteredPeers = filterTargetPeers(peers, params, field)
			if len(filteredPeers) == 0 {
				// 타겟이 설정되었지만 현재 연결된 peer 중 매칭되는 것이 없음
				log.Debug("BYZ: No matching peers for configured targets")
			}

		default:
			log.Debug("BYZ: unknown target for byzantine attack", "target", field.Target)
		}
	}

	// Note: If both direction and targets are specified, they work as OR condition
	// - direction=Send drops all messages
	// - targets filters specific peers (if not set and direction is set, all targets will be received message)
	// Either condition triggers the attack
	if shouldDrop || len(filteredPeers) > 0 {
		logAttackExecution(attack, blockNum, "params", params)
		hook.MarkAttackExecuted(attack.UID, blockNum)
	}

	return filteredPeers, shouldDrop
}

// checkMessageDirection checks if message should be dropped based on direction
func checkMessageDirection(field btypes.Field) bool {
	dirValue, ok := field.Value.(uint64)
	if !ok {
		log.Warn("BYZ: invalid direction value type", "value", field.Value)
		return false
	}

	if dirValue == uint64(btypes.MessageDirectionSend) || dirValue == uint64(btypes.MessageDirectionBoth) {
		return true
	}

	return false
}

// filterTargetPeers filters peers based on target configuration
func filterTargetPeers(peers []*ethPeer, params *btypes.MessagePolicyParams, field btypes.Field) []*ethPeer {
	var filtered []*ethPeer

	for _, peer := range peers {
		if params.IsTargetPeer(peer.Info().Enode) {
			filtered = append(filtered, peer)
		}
	}

	return filtered
}

// processOmitAttack handles omit message attacks
func processOmitAttack(attack *btypes.ExecutableAttack, block *types.Block, h *handler, hook btypes.ConsensusHook) *types.Block {
	modifiedBlock := h.createBlockWithMissingSeals(block, attack.OmitParams.Cmd)

	if modifiedBlock == nil {
		log.Warn("BYZ: Failed to create modified block", "cmd", attack.OmitParams.Cmd)
		return nil
	}

	logAttackExecution(attack, block.NumberU64(),
		"cmd", attack.OmitParams.Cmd,
		"original_hash", block.Hash(),
		"modified_hash", modifiedBlock.Hash(),
		"params", attack.OmitParams)

	hook.MarkAttackExecuted(attack.UID, block.NumberU64())
	return modifiedBlock
}

// processFakeAttack handles fake message attacks
func processFakeAttack(attack *btypes.ExecutableAttack, block *types.Block, h *handler, hook btypes.ConsensusHook) *types.Block {
	modifiedBlock := h.createBlockWithFakeSeals(block, attack.FakeParams)

	if modifiedBlock == nil {
		log.Warn("BYZ: Failed to create modified block", "fakeFields", attack.FakeParams.Fields)
		return nil
	}

	logAttackExecution(attack, block.NumberU64(),
		"original_hash", block.Hash(),
		"modified_hash", modifiedBlock.Hash(),
		"params", attack.FakeParams)

	hook.MarkAttackExecuted(attack.UID, block.NumberU64())
	return modifiedBlock
}

// logAttackExecution logs Byzantine attack execution with consistent format
func logAttackExecution(attack *btypes.ExecutableAttack, blockNum uint64, details ...interface{}) {
	// Build log context with standard fields
	logContext := []interface{}{
		"name", attack.NAME,
		"uid", attack.UID,
		"seq", blockNum,
	}

	// Append additional details
	logContext = append(logContext, details...)

	log.Info("BYZ: byzantine attack triggered", logContext...)
}

// IsAttackEnabled checks if a specific attack type is enabled
func IsAttackEnabled(attacks map[btypes.AttackType]*btypes.ExecutableAttack, attackType btypes.AttackType) bool {
	attack, exists := attacks[attackType]
	if attack == nil || !exists {
		return false
	}

	existAttackParams := false
	switch attackType {
	case btypes.AttackTypeMessagePolicy:
		if attack.MessagePolicyParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeTamperedMessage:
		if attack.TamperParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeFakeMessage:
		if attack.FakeParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeOmitMessage:
		if attack.OmitParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeRoleSpoofed:
		if attack.RoleSpoofParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeReplay:
		if attack.ReplayParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeStoreMessage:
		if attack.StoreMessageParams != nil {
			existAttackParams = true
		}
	case btypes.AttackTypeDos:
		if attack.DosParams != nil {
			existAttackParams = true
		}
	default:
		return false
	}
	return existAttackParams
}
