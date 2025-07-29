package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/common"
)

// MessagePolicyParams handles parsing for message policy parameters
type MessagePolicyParams struct {
	Code    MessageCode      `json:"code"`
	Fields  []Field          `json:"fields"`
	Targets []common.Address `json:"targets,omitempty"`
}

var _ AttackParamsParser = (*MessagePolicyParams)(nil)

// HasMessageCode checks if the attack applies to a specific message code
func (p *MessagePolicyParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

// ShouldBlockSend checks if sending should be blocked
func (p *MessagePolicyParams) ShouldBlockSend() bool {
	for _, field := range p.Fields {
		if field.Target == "policy.direction" {
			if v, ok := field.Value.(float64); ok {
				return v == 1 || v == 3
			}
		}
	}
	return false
}

// ShouldBlockReceive checks if receiving should be blocked
func (p *MessagePolicyParams) ShouldBlockReceive() bool {
	for _, field := range p.Fields {
		if field.Target == "policy.direction" {
			if v, ok := field.Value.(float64); ok {
				return v == 2 || v == 3
			}
		}
	}
	return false
}

// IsTargeted checks if a specific address is targeted
func (p *MessagePolicyParams) IsTargeted(addr common.Address) bool {
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
func (p *MessagePolicyParams) GetBlockedTargets(valSet []common.Address) []common.Address {
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

func (p *MessagePolicyParams) UnmarshalJSON(data []byte) error {
	type Alias MessagePolicyParams
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

func (p *MessagePolicyParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &MessagePolicyParams{}
	params.Code = ParseMessageCode(raw["code"])

	// Parse fields (supports both 'fields' and 'tamperFields' for backward compatibility)
	var fields []interface{}
	if f, ok := raw["fields"].([]interface{}); ok {
		fields = f
	}

	if fields != nil {
		params.Fields = make([]Field, 0, len(fields))
		for _, field := range fields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				targetStr := fmt.Sprintf("%v", fieldMap["target"])
				f := Field{
					Target: targetStr,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, f)
			}
		}
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

func (p *MessagePolicyParams) ParseJSON(data []byte) (interface{}, error) {
	params := &MessagePolicyParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *MessagePolicyParams) Validate(params interface{}) error {
	messagePolicyParams, ok := params.(*MessagePolicyParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(messagePolicyParams.Code) {
		return fmt.Errorf("invalid message code: %d", messagePolicyParams.Code)
	}

	// Validate Fields
	for i, field := range messagePolicyParams.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}
		// Value can be empty/nil as it might be set dynamically
	}

	return nil
}

// Field represents a generic field with target and value
type Field struct {
	Target string      `json:"target"` // e.g., "Header.Coinbase"
	Value  interface{} `json:"value"`  // Value for the field
}

// ValueToUint64 converts the Value field to uint64 if possible
func (f *Field) ValueToUint64() (uint64, error) {
	switch v := f.Value.(type) {
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
func (f *Field) ValueToHash() (common.Hash, error) {
	switch v := f.Value.(type) {
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

// ValueHexToBytes converts a hex string in the Value field (e.g., "0x1234") to []byte
func (f *Field) ValueHexToBytes() ([]byte, error) {
	str, ok := f.Value.(string)
	if !ok {
		return nil, fmt.Errorf("value is not a string: %T", f.Value)
	}

	// Remove "0x" or "0X" prefix if present
	trimmed := strings.TrimPrefix(strings.ToLower(str), "0x")

	// Hex strings must have even length
	if len(trimmed)%2 != 0 {
		return nil, fmt.Errorf("hex string has odd length: %d", len(trimmed))
	}

	// Decode the hex string to []byte
	bz, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	return bz, nil
}

// TamperAttackParams handles parsing for tamper attack parameters
type TamperAttackParams struct {
	Code             MessageCode      `json:"code"`
	Fields           []Field          `json:"fields"`
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

	// Parse fields (supports both 'fields' and 'tamperFields' for backward compatibility)
	var fields []interface{}
	if f, ok := raw["fields"].([]interface{}); ok {
		fields = f
	} else if f, ok := raw["tamperFields"].([]interface{}); ok {
		fields = f
	}

	if fields != nil {
		params.Fields = make([]Field, 0, len(fields))
		for _, field := range fields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				targetStr := fmt.Sprintf("%v", fieldMap["target"])
				f := Field{
					Target: targetStr,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, f)
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

	for i, field := range tamperParams.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}

		// Check if field.Value is a hex string and validate its length
		if strValue, ok := field.Value.(string); ok {
			// Remove 0x prefix if present
			hexStr := strings.TrimPrefix(strValue, "0x")

			// Check if it's a valid hex string
			if _, err := hex.DecodeString(hexStr); err != nil {
				return fmt.Errorf("field[%d] value is not a valid hex string: %v", i, err)
			}

			// Check if it's 32 bytes (64 hex characters)
			if len(hexStr) != 64 {
				return fmt.Errorf("field[%d] value must be 32 bytes (64 hex characters), got %d", i, len(hexStr)/2)
			}
		}
	}

	return nil
}

// FakeAttackParams handles parsing for fake attack parameters
type FakeAttackParams struct {
	Code    MessageCode      `json:"code"`
	Fields  []Field          `json:"fields,omitempty"`
	Targets []common.Address `json:"targets,omitempty"`
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

	// Parse fields (supports both 'fields' and 'fakeMessage' for backward compatibility)
	var fields []interface{}
	if f, ok := raw["fields"].([]interface{}); ok {
		fields = f
	} else if f, ok := raw["fakeMessage"].([]interface{}); ok {
		fields = f
	}

	if fields != nil {
		params.Fields = make([]Field, 0, len(fields))
		for _, field := range fields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				// Support both 'target' and 'fakeTarget' for backward compatibility
				var target string
				if t, ok := fieldMap["target"]; ok {
					target = fmt.Sprintf("%v", t)
				} else if t, ok := fieldMap["fakeTarget"]; ok {
					target = fmt.Sprintf("%v", t)
				}

				f := Field{
					Target: target,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, f)
			}
		}
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

	// Validate Fields
	for i, field := range fakeParams.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}
		// Value can be empty/nil as it might be set dynamically
	}

	return nil
}

// OmitAttackParams handles parsing for omit attack parameters
type OmitAttackParams struct {
	Code    MessageCode      `json:"code"`
	Cmd     uint64           `json:"cmd"`
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

	// Cmd and Option can be 0
	return nil
}

// RoleSpoofAttackParams handles parsing for role spoof attack parameters
type RoleSpoofAttackParams struct {
	Code    MessageCode      `json:"code"`
	Fields  []Field          `json:"fields,omitempty"`
	Targets []common.Address `json:"targets,omitempty"`
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

	// Parse fields
	var fields []interface{}
	if f, ok := raw["fields"].([]interface{}); ok {
		fields = f
	}

	if fields != nil {
		params.Fields = make([]Field, 0, len(fields))
		for _, field := range fields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				targetStr := fmt.Sprintf("%v", fieldMap["target"])

				f := Field{
					Target: targetStr,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, f)
			}
		}
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

	// Validate Fields
	for i, field := range roleSpoofParams.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}
		// Value can be empty/nil as it might be set dynamically
	}

	return nil
}

// ReplayAttackParams handles parsing for replay attack parameters
type ReplayAttackParams struct {
	Code            MessageCode      `json:"code"`
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

	if !ValidateMessageCode(replayParams.Code) {
		return fmt.Errorf("invalid message code: %d", replayParams.Code)
	}

	return nil
}

// StoreAttackParams handles parsing for replay attack parameters
type StoreAttackParams struct {
	Code MessageCode `json:"code"`
}

var _ AttackParamsParser = (*StoreAttackParams)(nil)

func (p *StoreAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *StoreAttackParams) UnmarshalJSON(data []byte) error {
	type Alias StoreAttackParams
	aux := &struct {
		*Alias
		Targets []string `json:"targets,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}

func (p *StoreAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &StoreAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	return params, nil
}

func (p *StoreAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &StoreAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *StoreAttackParams) Validate(params interface{}) error {
	storeParams, ok := params.(*StoreAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}

	if !ValidateMessageCode(storeParams.Code) {
		return fmt.Errorf("invalid message code: %d", storeParams.Code)
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

// DosAttackParams represents parameters for DoS attack
type DosAttackParams struct {
	Code    MessageCode      `json:"code"`
	Fields  []Field          `json:"fields"` // DOS type configuration
	Targets []common.Address `json:"targets"`
}

var _ AttackParamsParser = (*DosAttackParams)(nil)

// HasMessageCode checks if the given message code matches
func (p *DosAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *DosAttackParams) UnmarshalJSON(data []byte) error {
	type Alias DosAttackParams
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

// Parse parses raw parameters into DosAttackParams
func (p *DosAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &DosAttackParams{}
	params.Code = ParseMessageCode(raw["code"])

	// Parse fields - handle both []interface{} and []map[string]interface{}
	switch fields := raw["fields"].(type) {
	case []interface{}:
		params.Fields = make([]Field, 0, len(fields))
		for _, f := range fields {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				targetStr := fmt.Sprintf("%v", fieldMap["target"])
				field := Field{
					Target: targetStr,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, field)
			}
		}
	case []map[string]interface{}:
		params.Fields = make([]Field, 0, len(fields))
		for _, fieldMap := range fields {
			targetStr := fmt.Sprintf("%v", fieldMap["target"])
			field := Field{
				Target: targetStr,
				Value:  fieldMap["value"],
			}
			params.Fields = append(params.Fields, field)
		}
	default:
		log.Error("[BYZ] invalid field type", "type", reflect.TypeOf(raw["fields"]))
	}

	params.Targets = ParseTargets(raw["targets"])

	return params, nil
}

// ParseJSON parses JSON data into DosAttackParams
func (p *DosAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &DosAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

// Validate validates DosAttackParams
func (p *DosAttackParams) Validate(params interface{}) error {
	dosParams, ok := params.(*DosAttackParams)
	if !ok {
		return fmt.Errorf("invalid params type: expected *DosAttackParams")
	}

	if !ValidateMessageCode(dosParams.Code) {
		return fmt.Errorf("invalid message code: %d", dosParams.Code)
	}

	// Validate fields
	for _, field := range dosParams.Fields {
		// Validate target
		if field.Target != "valid" && field.Target != "invalid" {
			return fmt.Errorf("invalid target: %s, must be 'valid' or 'invalid'", field.Target)
		}

		// Validate value
		switch v := field.Value.(type) {
		case string:
			validValues := []string{"sequence", "round", "random", "signature", "blockHash"}
			valid := false
			for _, vv := range validValues {
				if v == vv {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid value: %s", v)
			}
		default:
			// Allow other types for flexibility
		}
	}

	return nil
}
