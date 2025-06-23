package attack

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// RoleType represents different node roles in QBFT consensus
type RoleType string

const (
	RoleTypeProposer  RoleType = "proposer"
	RoleTypeValidator RoleType = "validator"
	RoleTypeObserver  RoleType = "observer"
)

// RoleSpoofAttack implements an attack where the node spoofs another node's role
type RoleSpoofAttack struct {
	*baseAttack

	// Attack configuration
	messageCode   types.MessageCode
	spoofRole     string
	spoofedNodeID common.Address // Node ID to spoof (optional)
	fakeMessage   []byte         // Fake message content for the spoofed role
}

// NewRoleSpoofAttack creates a new role spoof attack
func NewRoleSpoofAttack(config *types.AttackConfig) (types.Attack, error) {
	//sequence := config.Sequence
	//round := config.Round

	// Parse message code
	codeStr, ok := config.Params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}
	messageCode := types.ParseMessageCode(codeStr)

	// Parse spoof role
	spoofRole, ok := config.Params["spoofRole"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'spoofRole' parameter")
	}

	// Parse spoofed node ID (optional)
	var spoofedNodeID common.Address
	if nodeID, ok := config.Params["spoofedNodeID"].(common.Address); ok {
		spoofedNodeID = nodeID
	} else if nodeIDStr, ok := config.Params["spoofedNodeID"].(string); ok {
		if common.IsHexAddress(nodeIDStr) {
			spoofedNodeID = common.HexToAddress(nodeIDStr)
		}
	}

	// Parse fake message (required)
	var fakeMessage []byte
	if msg, ok := config.Params["fakeMessage"].([]byte); ok {
		fakeMessage = msg
	} else if msg, ok := config.Params["fakeMessage"].(string); ok {
		// Allow hex string input
		if common.IsHexAddress(msg) || len(msg) > 2 && msg[:2] == "0x" {
			fakeMessage = common.FromHex(msg)
		} else {
			fakeMessage = []byte(msg)
		}
	} else if msg, ok := config.Params["fakeMessage"].([]interface{}); ok {
		// Convert from interface array
		for _, b := range msg {
			if byteVal, ok := b.(uint8); ok {
				fakeMessage = append(fakeMessage, byteVal)
			} else if byteVal, ok := b.(float64); ok {
				fakeMessage = append(fakeMessage, uint8(byteVal))
			}
		}
	}

	if len(fakeMessage) == 0 {
		return nil, fmt.Errorf("missing or empty 'fakeMessage' parameter")
	}

	// Parse targets
	var targets []common.Address
	if targetList, ok := config.Params["targets"].([]common.Address); ok {
		targets = targetList
	} else if targetList, ok := config.Params["targets"].([]interface{}); ok {
		for _, target := range targetList {
			if addr, ok := target.(common.Address); ok {
				targets = append(targets, addr)
			} else if addrStr, ok := target.(string); ok {
				if common.IsHexAddress(addrStr) {
					targets = append(targets, common.HexToAddress(addrStr))
				}
			}
		}
	}

	// Validate role spoofing for message type
	if err := validateRoleSpoofing(string(messageCode), spoofRole); err != nil {
		return nil, fmt.Errorf("invalid role spoofing: %w", err)
	}

	attack := &RoleSpoofAttack{
		// TODO:
		// should refactoring
		baseAttack:    nil,
		messageCode:   messageCode,
		spoofRole:     spoofRole,
		spoofedNodeID: spoofedNodeID,
		fakeMessage:   fakeMessage,
	}

	return attack, nil
}

// Type returns the attack type
func (r *RoleSpoofAttack) Type() types.AttackType {
	return ""
}

// Execute executes the role spoof attack
func (r *RoleSpoofAttack) Execute(ctx *types.AttackContext) error {
	// Check if this is the target message type
	//if ctx.Message == nil || ctx.Message.Code() != r.messageCode {
	//	return nil
	//}

	log.Info("Executing role spoof attack",
		//"name", r.Name(),
		//"sequence", ctx.Sequence,
		//"round", ctx.Round,
		"messageType", r.messageCode,
		"spoofRole", r.spoofRole,
		"spoofedNodeID", r.spoofedNodeID.Hex(),
		"fakeMessageSize", len(r.fakeMessage))
	//"targets", len(r.GetTargets()))

	// Log role spoof details for debugging
	log.Debug("Role spoof attack details",
		"originalRole", r.determineCurrentRole(ctx),
		"spoofRole", r.spoofRole,
		"messageHex", common.Bytes2Hex(r.fakeMessage))

	// Validate the spoofing attempt
	if err := r.validateSpoofing(ctx); err != nil {
		log.Error("Invalid role spoofing attempt", "error", err)
		//r.RecordExecution(false, err)
		return err
	}

	// The actual role spoofing will be handled by the interceptor
	// This just marks the execution and provides the spoofing parameters
	//r.RecordExecution(true, nil)

	return nil
}

// GetMessageCode returns the target message code
func (r *RoleSpoofAttack) GetMessageCode() types.MessageCode {
	return r.messageCode
}

// GetSpoofedRole returns the role to spoof
func (r *RoleSpoofAttack) GetSpoofedRole() string {
	return r.spoofRole
}

// GetSpoofedNodeID returns the node ID to spoof
func (r *RoleSpoofAttack) GetSpoofedNodeID() common.Address {
	return r.spoofedNodeID
}

// GetFakeMessage returns the fake message content
func (r *RoleSpoofAttack) GetFakeMessage() []byte {
	return r.fakeMessage
}

// SetSpoofedNodeID sets the node ID to spoof
func (r *RoleSpoofAttack) SetSpoofedNodeID(nodeID common.Address) {
	r.spoofedNodeID = nodeID
}

// validateSpoofing validates the spoofing attempt
func (r *RoleSpoofAttack) validateSpoofing(ctx *types.AttackContext) error {
	// Check if we're trying to spoof our own role
	currentRole := r.determineCurrentRole(ctx)
	if currentRole == r.spoofRole {
		return fmt.Errorf("cannot spoof own role: %s", currentRole)
	}

	// Check if spoofed node ID is valid (if specified)
	if r.spoofedNodeID != (common.Address{}) {
		// Verify the spoofed node is in the validator set
		found := false
		//for _, validator := range ctx.Validators {
		//	if validator == r.spoofedNodeID {
		//		found = true
		//		break
		//	}
		//}
		if !found {
			return fmt.Errorf("spoofed node ID not in validator set: %s", r.spoofedNodeID.Hex())
		}
	}

	// Validate fake message
	if len(r.fakeMessage) == 0 {
		return fmt.Errorf("fake message cannot be empty")
	}

	return nil
}

// determineCurrentRole determines the current node's role in the consensus
func (r *RoleSpoofAttack) determineCurrentRole(ctx *types.AttackContext) string {
	//if ctx.IsProposer {
	//	return string(RoleTypeProposer)
	//}

	// Check if we're in the validator set
	//for _, validator := range ctx.Validators {
	//	if validator == ctx.Self {
	//		return string(RoleTypeValidator)
	//	}
	//}

	return string(RoleTypeObserver)
}

// ShouldSpoofInMessage checks if spoofing should be applied to a specific message
func (r *RoleSpoofAttack) ShouldSpoofInMessage(msgType string, fromRole string) bool {
	// Only spoof if message type matches and we're spoofing the right role
	return msgType == string(r.messageCode) && fromRole == r.spoofRole
}

// CreateSpoofedMessage creates a spoofed message with the fake content
func (r *RoleSpoofAttack) CreateSpoofedMessage(originalMessage types.Message) (types.Message, error) {
	// This is a placeholder for the actual message spoofing logic
	// In a real implementation, this would:
	// 1. Clone the original message structure
	// 2. Replace the sender with spoofed node ID
	// 3. Replace the content with fake message
	// 4. Re-sign with spoofed credentials (if available)

	log.Debug("Creating spoofed message",
		"originalFrom", originalMessage.From().Hex(),
		"spoofedFrom", r.spoofedNodeID.Hex(),
		"fakeMessageSize", len(r.fakeMessage))

	// For now, return the original message
	// Real implementation would create a properly spoofed message
	return originalMessage, nil
}

// validateRoleSpoofing validates if role spoofing is valid for the message type
func validateRoleSpoofing(messageCode string, spoofedRole string) error {
	switch messageCode {
	case "PrePrepare":
		// Only proposers can send PrePrepare messages
		if spoofedRole != string(RoleTypeProposer) {
			return fmt.Errorf("PrePrepare messages can only be spoofed as proposer role")
		}
	case "Prepare", "Commit":
		// Both proposers and validators can send Prepare/Commit messages
		if spoofedRole != string(RoleTypeProposer) && spoofedRole != string(RoleTypeValidator) {
			return fmt.Errorf("Prepare/Commit messages can only be spoofed as proposer or validator role")
		}
	case "RoundChange":
		// All roles can send RoundChange messages
		// No restriction
	case "Propagation":
		// Usually sent by proposers
		if spoofedRole != string(RoleTypeProposer) {
			return fmt.Errorf("Propagation messages are typically spoofed as proposer role")
		}
	default:
		return fmt.Errorf("role spoofing not supported for message type: %s", messageCode)
	}
	return nil
}

// GetValidRolesForMessage returns valid roles that can send a specific message type
func GetValidRolesForMessage(messageCode string) []string {
	switch messageCode {
	case "PrePrepare":
		return []string{string(RoleTypeProposer)}
	case "Prepare", "Commit":
		return []string{string(RoleTypeProposer), string(RoleTypeValidator)}
	case "RoundChange":
		return []string{string(RoleTypeProposer), string(RoleTypeValidator), string(RoleTypeObserver)}
	case "Propagation":
		return []string{string(RoleTypeProposer)}
	default:
		return []string{}
	}
}

func (r *RoleSpoofAttack) ID() string                                   { return "" }
func (r *RoleSpoofAttack) SetID(id uint64)                              {}
func (r *RoleSpoofAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (r *RoleSpoofAttack) Configure(params interface{}) error           { return nil }
func (r *RoleSpoofAttack) Validate() error                              { return nil }
func (r *RoleSpoofAttack) Reset() error                                 { return nil }
func (r *RoleSpoofAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (r *RoleSpoofAttack) FromJSON(data []byte) error                   { return nil }

func (s *RoleSpoofAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *RoleSpoofAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *RoleSpoofAttack) RequiresData() []types.DataRequirement                     { return nil }

// RoleSpoofAttackFactory creates role spoof attacks
func RoleSpoofAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewRoleSpoofAttack(config)
}
