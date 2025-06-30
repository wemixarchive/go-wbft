package types

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

// SilentAttackParams handles parsing for silent attack parameters
type SilentAttackParams struct {
	Code      MessageCode
	Direction uint64
	Targets   []common.Address
}

func (p *SilentAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &SilentAttackParams{}
	params.Code = ParseMessageCode(raw["code"])

	if direction, ok := raw["direction"].(float64); ok {
		params.Direction = uint64(direction)
	} else if direction, ok := raw["direction"].(int); ok {
		params.Direction = uint64(direction)
	} else if direction, ok := raw["direction"].(uint); ok {
		params.Direction = uint64(direction)
	} else if direction, ok := raw["direction"].(uint64); ok {
		params.Direction = direction
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *SilentAttackParams) Validate(params interface{}) error {
	silentParams, ok := params.(*SilentAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if silentParams.Direction > 3 {
		return fmt.Errorf("invalid direction: %d (must be 0-3)", silentParams.Direction)
	}

	if !ValidateMessageCode(silentParams.Code) {
		return fmt.Errorf("invalid message code: %d", silentParams.Code)
	}

	return nil
}

// TamperAttackParams handles parsing for tamper attack parameters
type TamperAttackParams struct {
	Code             MessageCode
	TamperFields     []TamperField
	WithValidMessage bool
	Delay            uint64
	Targets          []common.Address
}

// TamperField represents a field to be tampered with in a message
type TamperField struct {
	Target string      `json:"target"` // e.g., "Proposal.Header.Coinbase"
	Value  interface{} `json:"value"`  // New value for the field
}

func (p *TamperAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &TamperAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	// Parse tamperFields
	if tamperFields, ok := raw["tamperFields"].([]interface{}); ok {
		params.TamperFields = make([]TamperField, 0, len(tamperFields))
		for _, field := range tamperFields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				tamperField := TamperField{
					Target: fmt.Sprintf("%v", fieldMap["target"]),
					Value:  fieldMap["value"],
				}
				params.TamperFields = append(params.TamperFields, tamperField)
			}
		}
	}

	if withValid, ok := raw["withValidMessage"].(bool); ok {
		params.WithValidMessage = withValid
	}

	if delay, ok := raw["delay"].(float64); ok {
		params.Delay = uint64(delay)
	} else if delay, ok := raw["delay"].(int); ok {
		params.Delay = uint64(delay)
	} else if delay, ok := raw["delay"].(int64); ok {
		params.Delay = uint64(delay)
	} else if delay, ok := raw["delay"].(uint64); ok {
		params.Delay = delay
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *TamperAttackParams) Validate(params interface{}) error {
	tamperParams, ok := params.(*TamperAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if len(tamperParams.TamperFields) == 0 && !tamperParams.WithValidMessage {
		return fmt.Errorf("tamper attack must have either tamperFields or withValidMessage=true")
	}

	for i, field := range tamperParams.TamperFields {
		if field.Target == "" {
			return fmt.Errorf("tamperField[%d] target is empty", i)
		}
	}

	if !ValidateMessageCode(tamperParams.Code) {
		return fmt.Errorf("invalid message code: %d", tamperParams.Code)
	}

	return nil
}

// FakeAttackParams handles parsing for fake attack parameters
type FakeAttackParams struct {
	Code        MessageCode
	FakeMessage []byte
	Targets     []common.Address
}

func (p *FakeAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &FakeAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if fakeMsg, ok := raw["fakeMessage"].(string); ok {
		params.FakeMessage = []byte(fakeMsg)
	} else if fakeMsg, ok := raw["fakeMessage"].([]byte); ok {
		params.FakeMessage = fakeMsg
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *FakeAttackParams) Validate(params interface{}) error {
	fakeParams, ok := params.(*FakeAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(fakeParams.Code) {
		return fmt.Errorf("invalid message code: %d", fakeParams.Code)
	}

	// FakeMessage can be empty as it might be generated later
	return nil
}

// OmitAttackParams handles parsing for omit attack parameters
type OmitAttackParams struct {
	Code    MessageCode
	Cmd     uint64
	Cnt     uint64
	Targets []common.Address
}

func (p *OmitAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &OmitAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if cmd, ok := raw["cmd"].(float64); ok {
		params.Cmd = uint64(cmd)
	} else if cmd, ok := raw["cmd"].(int); ok {
		params.Cmd = uint64(cmd)
	}

	if cnt, ok := raw["cnt"].(float64); ok {
		params.Cnt = uint64(cnt)
	} else if cnt, ok := raw["cnt"].(int); ok {
		params.Cnt = uint64(cnt)
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *OmitAttackParams) Validate(params interface{}) error {
	omitParams, ok := params.(*OmitAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(omitParams.Code) {
		return fmt.Errorf("invalid message code: %d", omitParams.Code)
	}

	// Cmd and Cnt can be 0
	return nil
}

// RoleSpoofAttackParams handles parsing for role spoof attack parameters
type RoleSpoofAttackParams struct {
	Code        MessageCode
	FakeMessage []byte
	Targets     []common.Address
}

func (p *RoleSpoofAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &RoleSpoofAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if fakeMsg, ok := raw["fakeMessage"].(string); ok {
		params.FakeMessage = []byte(fakeMsg)
	} else if fakeMsg, ok := raw["fakeMessage"].([]byte); ok {
		params.FakeMessage = fakeMsg
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *RoleSpoofAttackParams) Validate(params interface{}) error {
	roleSpoofParams, ok := params.(*RoleSpoofAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(roleSpoofParams.Code) {
		return fmt.Errorf("invalid message code: %d", roleSpoofParams.Code)
	}

	// FakeMessage can be empty
	return nil
}

// ReplayAttackParams handles parsing for replay attack parameters
type ReplayAttackParams struct {
	Code            MessageCode
	OriSequence     uint64
	OriRound        uint64
	UseOriginalView bool
	Targets         []common.Address
}

func (p *ReplayAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &ReplayAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if oriSeq, ok := raw["ori_sequence"].(float64); ok {
		params.OriSequence = uint64(oriSeq)
	} else if oriSeq, ok := raw["ori_sequence"].(int); ok {
		params.OriSequence = uint64(oriSeq)
	}

	if oriRound, ok := raw["ori_round"].(float64); ok {
		params.OriRound = uint64(oriRound)
	} else if oriRound, ok := raw["ori_round"].(int); ok {
		params.OriRound = uint64(oriRound)
	}

	if useOrigView, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = useOrigView
	}

	params.Targets = parseTargets(raw["targets"])

	return params, nil
}

func (p *ReplayAttackParams) Validate(params interface{}) error {
	replayParams, ok := params.(*ReplayAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if replayParams.OriSequence == 0 {
		return fmt.Errorf("original sequence must be specified")
	}

	if !ValidateMessageCode(replayParams.Code) {
		return fmt.Errorf("invalid message code: %d", replayParams.Code)
	}

	return nil
}

// Helper function to parse targets
func parseTargets(rawTargets interface{}) []common.Address {
	var targets []common.Address

	switch v := rawTargets.(type) {
	case []interface{}:
		for _, target := range v {
			if addr, ok := target.(string); ok && common.IsHexAddress(addr) {
				targets = append(targets, common.HexToAddress(addr))
			}
		}
	case []string:
		for _, addr := range v {
			if common.IsHexAddress(addr) {
				targets = append(targets, common.HexToAddress(addr))
			}
		}
	case []common.Address:
		targets = v
	}

	return targets
}
