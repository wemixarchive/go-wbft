package eth

import (
	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// BlockAnnouncement represents a block announcement
type BlockAnnouncement struct {
	Hash   common.Hash
	Number uint64
}

// CheckByzantinePropagationAttacks checks multiple block announcements for Byzantine attacks
// Returns nil if any attack is detected, otherwise returns unaffected announcements
func CheckByzantinePropagationAttacks(
	backend interface{ ByzantineHook() btypes.ConsensusHook },
	announcements []BlockAnnouncement, // BlockAnnouncement는 hash와 number를 포함하는 구조체
	peerID string,
) []BlockAnnouncement {
	hook := backend.ByzantineHook()
	if hook == nil {
		return announcements
	}

	validAnnouncements := make([]BlockAnnouncement, 0, len(announcements))

	for _, ann := range announcements {
		if CheckByzantinePropagationAttack(hook, ann.Number, ann.Hash, peerID) {
			// Attack detected, drop all announcements
			return nil
		}

		// Only add valid announcements (non-zero block numbers)
		if ann.Number > 0 {
			validAnnouncements = append(validAnnouncements, ann)
		}
	}

	return validAnnouncements
}

// CheckByzantinePropagationAttack checks for Byzantine attacks on block propagation messages
// Returns true if the message should be dropped due to an attack
func CheckByzantinePropagationAttack(
	hook btypes.ConsensusHook,
	blockNumber uint64,
	blockHash common.Hash,
	peerID string,
) bool {
	if hook == nil {
		return false
	}

	// Skip invalid block numbers
	if blockNumber == 0 {
		log.Warn("[BYZ] received new block announcement with zero block number",
			"peer", peerID,
			"hash", blockHash.Hex())
		return false
	}

	// Get executable attacks for this block number
	// Note: Using round 0 for propagation messages
	attacks := hook.GetExecutableAttacks(btypes.MessageCodePropagation, blockNumber, 0)

	// Check for message policy attack
	if at := attacks[btypes.AttackTypeMessagePolicy]; at != nil && at.MessagePolicyParams != nil {
		if shouldDropPropagationMessage(at.MessagePolicyParams, peerID) {
			log.Info("[BYZ] byzantine attack triggered",
				"name", at.NAME,
				"uid", at.UID,
				"seq", blockNumber,
				"params", at.MessagePolicyParams,
				"newBlockHash", blockHash.Hex(),
				"newBlockNumber", blockNumber,
				"peer", peerID)
			hook.MarkAttackExecuted(at.UID, blockNumber)
			return true
		}
	}

	return false
}

// shouldDropPropagationMessage checks if propagation message should be dropped
func shouldDropPropagationMessage(params *btypes.MessagePolicyParams, peerID string) bool {
	for _, field := range params.Fields {
		if field.Target == btypes.TargetMsgPolicyDirection {
			if val, ok := field.Value.(uint64); ok {
				direction := btypes.MessageDirection(val)
				if direction == btypes.MessageDirectionReceive ||
					direction == btypes.MessageDirectionBoth {
					return true
				}
			}
		}
	}
	return false
}
