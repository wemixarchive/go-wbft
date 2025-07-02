package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// SilentAttackParams handles parsing for silent attack parameters
type SilentAttackParams struct {
	Code      MessageCode      `json:"code"`
	Direction uint64           `json:"direction"`
	Targets   []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*SilentAttackParams)(nil)

// HasMessageCode checks if the attack applies to a specific message code
func (p *SilentAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

// ShouldBlockSend checks if sending should be blocked
func (p *SilentAttackParams) ShouldBlockSend() bool {
	return p.Direction == uint64(1) || p.Direction == uint64(3)
}

// ShouldBlockReceive checks if receiving should be blocked
func (p *SilentAttackParams) ShouldBlockReceive() bool {
	return p.Direction == uint64(2) || p.Direction == uint64(3)
}

// IsTargeted checks if a specific address is targeted
func (p *SilentAttackParams) IsTargeted(addr common.Address) bool {
	if len(p.Targets) == 0 {
		return true // No specific targets means all are targeted
	}

	for _, target := range p.Targets {
		if target == addr {
			return true
		}
	}
	return false
}

// GetBlockedTargets returns the list of addresses to block
func (p *SilentAttackParams) GetBlockedTargets(valSet []common.Address) []common.Address {
	if len(p.Targets) == 0 {
		// Block all validators
		return valSet
	}

	// Filter only targeted validators
	var blocked []common.Address
	for _, val := range valSet {
		if p.IsTargeted(val) {
			blocked = append(blocked, val)
		}
	}
	return blocked
}

func (p *SilentAttackParams) UnmarshalJSON(data []byte) error {
	type Alias SilentAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *SilentAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &SilentAttackParams{}
	params.Code = ParseMessageCode(raw["code"])

	switch v := raw["direction"].(type) {
	case float64:
		params.Direction = uint64(v)
	case int:
		params.Direction = uint64(v)
	case uint:
		params.Direction = uint64(v)
	case uint64:
		params.Direction = v
	case string:
		switch strings.ToLower(v) {
		case "send", "1":
			params.Direction = 1
		case "receive", "2":
			params.Direction = 2
		case "both", "3":
			params.Direction = 3
		default:
			return nil, fmt.Errorf("invalid direction: %s", v)
		}
	default:
		return nil, fmt.Errorf("invalid direction type: %T", v)
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *SilentAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &SilentAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *SilentAttackParams) Validate(params interface{}) error {
	silentParams, ok := params.(*SilentAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(silentParams.Code) {
		return fmt.Errorf("invalid message code: %d", silentParams.Code)
	}

	if silentParams.Direction > 3 {
		return fmt.Errorf("invalid direction: %d (must be 0-3)", silentParams.Direction)
	}

	return nil
}

// TamperField represents a field to be tampered with in a message
type TamperField struct {
	Target TamperTarget `json:"target"` // e.g., "Proposal.Header.Coinbase"
	Value  interface{}  `json:"value"`  // New value for the field
}

// ValueToUint64 converts the Value field to uint64 if possible
func (tf *TamperField) ValueToUint64() (uint64, error) {
	switch v := tf.Value.(type) {
	case float64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint64:
		return v, nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string to uint64: %w", err)
		}
		return parsed, nil
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("cannot parse json.Number to int64: %w", err)
		}
		return uint64(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported type for uint64 conversion: %T", v)
	}
}

// ValueToHash converts the Value field to common.Hash if possible
func (tf *TamperField) ValueToHash() (common.Hash, error) {
	switch v := tf.Value.(type) {
	case string:
		// Accepts "0x..." or raw hex
		str := strings.TrimPrefix(v, "0x")
		bytes, err := hex.DecodeString(str)
		if err != nil {
			return common.Hash{}, fmt.Errorf("invalid hex string for Hash: %w", err)
		}
		if len(bytes) != 32 {
			return common.Hash{}, fmt.Errorf("invalid hash length: expected 32 bytes, got %d", len(bytes))
		}
		return common.BytesToHash(bytes), nil

	case []byte:
		if len(v) != 32 {
			return common.Hash{}, fmt.Errorf("invalid byte slice length for Hash: %d", len(v))
		}
		return common.BytesToHash(v), nil

	case [32]byte:
		return common.BytesToHash(v[:]), nil

	case common.Hash:
		return v, nil

	default:
		return common.Hash{}, fmt.Errorf("unsupported type for Hash conversion: %T", v)
	}
}

// TamperAttackParams handles parsing for tamper attack parameters
type TamperAttackParams struct {
	Code             MessageCode      `json:"code"`
	TamperFields     []TamperField    `json:"tamperFields"`
	WithValidMessage bool             `json:"withValidMessage"`
	Delay            uint64           `json:"delay"`
	Targets          []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*TamperAttackParams)(nil)

func (p *TamperAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *TamperAttackParams) UnmarshalJSON(data []byte) error {
	type Alias TamperAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *TamperAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &TamperAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	// Parse tamperFieldsAdd commentMore actions
	if tamperFields, ok := raw["tamperFields"].([]interface{}); ok {
		params.TamperFields = make([]TamperField, 0, len(tamperFields))
		for _, field := range tamperFields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				tamperTargetStr := fmt.Sprintf("%v", fieldMap["target"])
				tamperField := TamperField{
					Target: TamperTarget(tamperTargetStr),
					Value:  fieldMap["value"],
				}
				params.TamperFields = append(params.TamperFields, tamperField)
			}
		}
	}

	if withValid, ok := raw["withValidMessage"].(bool); ok {
		params.WithValidMessage = withValid
	}

	switch v := raw["delay"].(type) {
	case float64:
		params.Delay = uint64(v)
	case int:
		params.Delay = uint64(v)
	case uint:
		params.Delay = uint64(v)
	case uint64:
		params.Delay = v
	default:
		params.Delay = 0
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *TamperAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &TamperAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *TamperAttackParams) Validate(params interface{}) error {
	tamperParams, ok := params.(*TamperAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(tamperParams.Code) {
		return fmt.Errorf("invalid message code: %d", tamperParams.Code)
	}

	for i, field := range tamperParams.TamperFields {
		if field.Target == "" {
			return fmt.Errorf("tamperField[%d] target is empty", i)
		}

		// Check if field.Value is a hex string and validate its length
		if strValue, ok := field.Value.(string); ok {
			// Remove 0x prefix if present
			hexStr := strings.TrimPrefix(strValue, "0x")

			// Check if it's a valid hex string
			if _, err := hex.DecodeString(hexStr); err != nil {
				return fmt.Errorf("tamperField[%d] value is not a valid hex string: %v", i, err)
			}

			// Check if it's 32 bytes (64 hex characters)
			if len(hexStr) != 64 {
				return fmt.Errorf("tamperField[%d] value must be 32 bytes (64 hex characters), got %d", i, len(hexStr)/2)
			}
		}
	}

	return nil
}

// FakeAttackParams handles parsing for fake attack parameters
type FakeAttackParams struct {
	Code        MessageCode      `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*FakeAttackParams)(nil)

func (p *FakeAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *FakeAttackParams) UnmarshalJSON(data []byte) error {
	type Alias FakeAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *FakeAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &FakeAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if fakeMsg, ok := raw["fakeMessage"].(string); ok {
		params.FakeMessage = []byte(fakeMsg)
	} else if fakeMsg, ok := raw["fakeMessage"].([]byte); ok {
		params.FakeMessage = fakeMsg
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *FakeAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &FakeAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
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
	Code    MessageCode      `json:"code"`
	Cmd     uint64           `json:"cmd"`
	Cnt     uint64           `json:"cnt"`
	Targets []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*OmitAttackParams)(nil)

func (p *OmitAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *OmitAttackParams) UnmarshalJSON(data []byte) error {
	type Alias OmitAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *OmitAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &OmitAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	switch v := raw["cmd"].(type) {
	case float64:
		params.Cmd = uint64(v)
	case int:
		params.Cmd = uint64(v)
	case uint:
		params.Cmd = uint64(v)
	case uint64:
		params.Cmd = v
	default:
		params.Cmd = 0
	}

	switch v := raw["cnt"].(type) {
	case float64:
		params.Cnt = uint64(v)
	case int:
		params.Cnt = uint64(v)
	case uint:
		params.Cnt = uint64(v)
	case uint64:
		params.Cnt = v
	default:
		params.Cnt = 0
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *OmitAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &OmitAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
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
	Code        MessageCode      `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*RoleSpoofAttackParams)(nil)

func (p *RoleSpoofAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *RoleSpoofAttackParams) UnmarshalJSON(data []byte) error {
	type Alias RoleSpoofAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *RoleSpoofAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &RoleSpoofAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if fakeMsg, ok := raw["fakeMessage"].(string); ok {
		params.FakeMessage = []byte(fakeMsg)
	} else if fakeMsg, ok := raw["fakeMessage"].([]byte); ok {
		params.FakeMessage = fakeMsg
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *RoleSpoofAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &RoleSpoofAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
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
	Code            MessageCode      `json:"code"`
	OriSequence     uint64           `json:"ori_sequence"`
	OriRound        uint64           `json:"ori_round"`
	UseOriginalView bool             `json:"useOriginalView"`
	Targets         []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*ReplayAttackParams)(nil)

func (p *ReplayAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *ReplayAttackParams) UnmarshalJSON(data []byte) error {
	type Alias ReplayAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Targets = make([]common.Address, 0, len(aux.Targets))
	for _, addr := range aux.Targets {
		if common.IsHexAddress(addr) {
			p.Targets = append(p.Targets, common.HexToAddress(addr))
		}
	}

	return nil
}

func (p *ReplayAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &ReplayAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	switch v := raw["ori_sequence"].(type) {
	case float64:
		params.OriSequence = uint64(v)
	case int:
		params.OriSequence = uint64(v)
	case uint:
		params.OriSequence = uint64(v)
	case uint64:
		params.OriSequence = v
	default:
		params.OriSequence = 0
	}

	switch v := raw["ori_round"].(type) {
	case float64:
		params.OriRound = uint64(v)
	case int:
		params.OriRound = uint64(v)
	case uint:
		params.OriRound = uint64(v)
	case uint64:
		params.OriRound = v
	default:
		params.OriRound = 0
	}

	if useOrigView, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = useOrigView
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *ReplayAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &ReplayAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
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

func ParseTargets(rawTargets interface{}) []common.Address {
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
