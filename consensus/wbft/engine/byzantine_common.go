package wbftengine

import (
	"fmt"
	"errors"
	
	btypes "github.com/ethereum/go-ethereum/byzantine/types"
)

func (e *Engine) GetByzantineHook() btypes.ConsensusHook {
	return e.backend.ByzantineHook()
}

func (e *Engine) GetExecutableByzantineAttacks(msgCode btypes.MessageCode, sequence, round uint64) (map[btypes.AttackType]*btypes.ExecutableAttack, error) {
	// get hook
	hook := e.GetByzantineHook()
	if hook == nil {
		return nil, fmt.Errorf("[BYZ] ByzantineHook is nil")
	}
	// get executable attacks
	attacks := hook.GetExecutableAttacks(msgCode, sequence, round)
	if attacks == nil {
		return nil, fmt.Errorf("[BYZ] ByzantineAttacks is nil")
	}

	return attacks, nil
}

func (e *Engine) MarkAttackExecuted(uid string, sequence uint64) error {
	if hook := e.GetByzantineHook(); hook != nil {
		return e.GetByzantineHook().MarkAttackExecuted(uid, sequence)
	}
	return errors.New("[BYZ] not found by byzantine hook")
}

func (e *Engine) ExistByzantineAttack(msgCode btypes.AttackType, attacks map[btypes.AttackType]*btypes.ExecutableAttack) *btypes.ExecutableAttack {
	for attackType, executableAttack := range attacks {
		if executableAttack != nil && executableAttack.Enabled {
			switch attackType {
			case btypes.AttackTypeMessagePolicy:
				if executableAttack.MessagePolicyParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeTamperedMessage:
				if executableAttack.TamperParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeFakeMessage:
				if executableAttack.FakeParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeOmitMessage:
				if executableAttack.OmitParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeRoleSpoofed:
				if executableAttack.RoleSpoofParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeReplay:
				if executableAttack.ReplayParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeStoreMessage:
				if executableAttack.StoreMessageParams != nil && attackType == msgCode {
					return executableAttack
				}
			case btypes.AttackTypeDos:
				if executableAttack.DosParams != nil && attackType == msgCode {
					return executableAttack
				}
			default:
				return nil
			}
		}
	}
	return nil
}
