package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/byzantine/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/bls"
)

// FakeBLSKey represents a fake BLS key with associated address
type FakeBLSKey struct {
	Address   common.Address
	SecretKey bls.SecretKey
	PublicKey bls.PublicKey
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

// MessagePolicyParams handles parsing for message policy parameters
type MessagePolicyParams struct {
	Code   MessageCode `json:"code"`
	Fields []Field     `json:"fields"`
}

var _ AttackParamsParser = (*MessagePolicyParams)(nil)

// HasMessageCode checks if the attack applies to a specific message code
func (p *MessagePolicyParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

// ShouldBlockSend checks if sending should be blocked
func (p *MessagePolicyParams) ShouldBlockSend() bool {
	for _, field := range p.Fields {
		if field.Target == TargetMsgPolicyDirection {
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
		if field.Target == TargetMsgPolicyDirection {
			if v, ok := field.Value.(float64); ok {
				return v == 2 || v == 3
			}
		}
	}
	return false
}

// IsTargeted checks if a specific address is targeted
func (p *MessagePolicyParams) IsTargeted(addr common.Address) bool {
	targets := p.GetTargets()
	if targets == nil || len(targets) == 0 {
		// No targets field means all are targeted
		return true
	}

	addrStr := addr.Hex()
	// Check if addr is in the targets list
	for _, target := range targets {
		// Check if target is an address
		if common.IsHexAddress(target) {
			if strings.EqualFold(target, addrStr) {
				return true
			}
		}
	}
	return false
}

func (p *MessagePolicyParams) IsTargetPeer(peer string) bool {
	targets := p.GetTargets()
	if targets == nil || len(targets) == 0 {
		// No targets means all peers are targeted
		return true
	}

	// Check each target
	for _, target := range targets {
		// Check if target is an enode URL or peer contains target
		if strings.Contains(peer, target) || strings.Contains(target, peer) {
			return true
		}
	}
	return false
}

// GetTargets extracts the targets from fields if present
func (p *MessagePolicyParams) GetTargets() []string {
	for _, field := range p.Fields {
		if field.Target == TargetMsgPolicyTargets {
			switch v := field.Value.(type) {
			case []string:
				return v
			case []common.Address:
				// Convert addresses to strings
				var strTargets []string
				for _, addr := range v {
					strTargets = append(strTargets, addr.Hex())
				}
				return strTargets
			case string:
				// Single target as string
				return []string{v}
			}
		}
	}
	return nil
}

// GetBlockedTargets returns the list of addresses to block
func (p *MessagePolicyParams) GetBlockedTargets(valSet []common.Address) []common.Address {
	targets := p.GetTargets()
	if targets != nil && len(targets) > 0 {
		// Convert string targets back to addresses if they are valid addresses
		var addrTargets []common.Address
		for _, target := range targets {
			if common.IsHexAddress(target) {
				addrTargets = append(addrTargets, common.HexToAddress(target))
			}
		}
		if len(addrTargets) > 0 {
			return addrTargets
		}
	}
	// No targets field means block all validators
	return valSet
}

func (p *MessagePolicyParams) UnmarshalJSON(data []byte) error {
	type Alias MessagePolicyParams
	return json.Unmarshal(data, (*Alias)(p))
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
				var fieldValue interface{}

				switch targetStr {
				case TargetMsgPolicyDirection, TargetMsgPolicyDelay:
					v, err := ParseToUint64(fieldMap["value"])
					if err != nil {
						return nil, fmt.Errorf("invalid value for %s: %w", targetStr, err)
					}
					fieldValue = v
				case TargetMsgPolicyTargets:
					// Special handling for targets field - use ParseTargets helper
					fieldValue = ParseTargets(fieldMap["value"])
				case TargetMsgPolicySendOriginal:
					fieldValue = fieldMap["value"]
				}

				f := Field{
					Target: targetStr,
					Value:  fieldValue,
				}
				params.Fields = append(params.Fields, f)
			}
		}
	}

	return params, nil
}

func (p *MessagePolicyParams) ParseJSON(data []byte) (interface{}, error) {
	params := &MessagePolicyParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *MessagePolicyParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	// Validate Fields
	for i, field := range p.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}

		// Special validation for targets field
		if field.Target == "targets" {
			switch v := field.Value.(type) {
			case []string:
				// Use validateTargets helper
				if err := validateTargets(v); err != nil {
					return fmt.Errorf("field[%d] targets: %w", i, err)
				}
			case []common.Address:
				// Use validateTargets helper
				if err := validateTargets(v); err != nil {
					return fmt.Errorf("field[%d] targets: %w", i, err)
				}
			case string:
				// Single target
				if err := validateTargets([]string{v}); err != nil {
					return fmt.Errorf("field[%d] targets: %w", i, err)
				}
			case nil:
				// Empty targets is allowed
			default:
				return fmt.Errorf("field[%d] targets: expected []string, []common.Address or string but got %T", i, v)
			}
		}
		// Other values can be empty/nil as they might be set dynamically
	}

	return nil
}

func (p *MessagePolicyParams) ValidateWith(params interface{}) error {
	messagePolicyParams, ok := params.(*MessagePolicyParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *MessagePolicyParams")
	}
	return messagePolicyParams.Validate()
}

// TamperAttackParams handles parsing for tamper attack parameters
type TamperAttackParams struct {
	Code   MessageCode `json:"code"`
	Fields []Field     `json:"fields"`
}

var _ AttackParamsParser = (*TamperAttackParams)(nil)

func (p *TamperAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *TamperAttackParams) UnmarshalJSON(data []byte) error {
	type Alias TamperAttackParams
	return json.Unmarshal(data, (*Alias)(p))
}

func (p *TamperAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &TamperAttackParams{}

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

	return params, nil
}

func (p *TamperAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &TamperAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *TamperAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	for i, field := range p.Fields {
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

func (p *TamperAttackParams) ValidateWith(params interface{}) error {
	tamperAttackParams, ok := params.(*TamperAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *TamperAttackParams")
	}
	return tamperAttackParams.Validate()
}

// FakeAttackParams handles parsing for fake attack parameters
type FakeAttackParams struct {
	Code   MessageCode `json:"code"`
	Fields []Field     `json:"fields,omitempty"`
}

var _ AttackParamsParser = (*FakeAttackParams)(nil)

func (p *FakeAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *FakeAttackParams) UnmarshalJSON(data []byte) error {
	type Alias FakeAttackParams
	return json.Unmarshal(data, (*Alias)(p))
}

func (p *FakeAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &FakeAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	// Parse fields (supports both 'fields' and 'fakeMessage' for backward compatibility)
	var fields []interface{}
	if f, ok := raw["fields"].([]interface{}); ok {
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
				}

				f := Field{
					Target: target,
					Value:  fieldMap["value"],
				}
				params.Fields = append(params.Fields, f)
			}
		}
	}

	return params, nil
}

func (p *FakeAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &FakeAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *FakeAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	// Validate Fields
	for i, field := range p.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}
		// Value can be empty/nil as it might be set dynamically
	}

	return nil
}

func (p *FakeAttackParams) ValidateWith(params interface{}) error {
	fakeAttackParams, ok := params.(*FakeAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *FakeAttackParams")
	}
	return fakeAttackParams.Validate()
}

// OmitAttackParams handles parsing for omit attack parameters
type OmitAttackParams struct {
	Code MessageCode `json:"code"`
	Cmd  uint64      `json:"cmd"`
}

var _ AttackParamsParser = (*OmitAttackParams)(nil)

func (p *OmitAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *OmitAttackParams) UnmarshalJSON(data []byte) error {
	type Alias OmitAttackParams
	return json.Unmarshal(data, (*Alias)(p))
}

func (p *OmitAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &OmitAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if cmd, err := ParseToUint64(raw["cmd"]); err == nil {
		params.Cmd = cmd
	} else {
		return nil, fmt.Errorf("invalid cmd value: %w", err)
	}

	return params, nil
}

func (p *OmitAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &OmitAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *OmitAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	if !(p.Code == MessageCodePrePrepare || p.Code == MessageCodePropagation) {
		return fmt.Errorf("invalid cmd value: %d, must be PrePrepare or Propagation", p.Cmd)
	}

	switch p.Code {
	case MessageCodePrePrepare:
		if !(p.Cmd == OmitCommandPrevPrepareSeal ||
			p.Cmd == OmitCommandPrevCommitSeal ||
			p.Cmd == OmitCommandRoundChange ||
			p.Cmd == OmitCommandPrepareMessage) {
			return fmt.Errorf("invalid cmd value: %d", p.Cmd)
		}
	case MessageCodePropagation:
		if !(p.Cmd == OmitCommandPrepareSeal ||
			p.Cmd == OmitCommandCommitSeal) {
			return fmt.Errorf("invalid cmd value: %d", p.Cmd)
		}
	default:
		return fmt.Errorf("unknown message code: %d", p.Code)
	}

	return nil
}

func (p *OmitAttackParams) ValidateWith(params interface{}) error {
	omitAttackParams, ok := params.(*OmitAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *OmitAttackParams")
	}
	return omitAttackParams.Validate()
}

// RoleSpoofAttackParams handles parsing for role spoof attack parameters
type RoleSpoofAttackParams struct {
	Code   MessageCode `json:"code"`
	Fields []Field     `json:"fields,omitempty"`
}

var _ AttackParamsParser = (*RoleSpoofAttackParams)(nil)

func (p *RoleSpoofAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *RoleSpoofAttackParams) UnmarshalJSON(data []byte) error {
	type Alias RoleSpoofAttackParams
	return json.Unmarshal(data, (*Alias)(p))
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

	return params, nil
}

func (p *RoleSpoofAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &RoleSpoofAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *RoleSpoofAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	// Validate Fields
	for i, field := range p.Fields {
		if field.Target == "" {
			return fmt.Errorf("field[%d] target is empty", i)
		}
		// Value can be empty/nil as it might be set dynamically
	}

	return nil
}

func (p *RoleSpoofAttackParams) ValidateWith(params interface{}) error {
	roleSpoofAttackParams, ok := params.(*RoleSpoofAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *RoleSpoofAttackParams")
	}
	return roleSpoofAttackParams.Validate()
}

// ReplayAttackParams handles parsing for replay attack parameters
type ReplayAttackParams struct {
	Code            MessageCode `json:"code"`
	UseOriginalView bool        `json:"useOriginalView"`
}

var _ AttackParamsParser = (*ReplayAttackParams)(nil)

func (p *ReplayAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *ReplayAttackParams) UnmarshalJSON(data []byte) error {
	type Alias ReplayAttackParams
	return json.Unmarshal(data, (*Alias)(p))
}

func (p *ReplayAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &ReplayAttackParams{}

	params.Code = ParseMessageCode(raw["code"])

	if useOrigView, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = useOrigView
	}

	return params, nil
}

func (p *ReplayAttackParams) ParseJSON(data []byte) (interface{}, error) {
	params := &ReplayAttackParams{}
	if err := json.Unmarshal(data, params); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *ReplayAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	return nil
}

func (p *ReplayAttackParams) ValidateWith(params interface{}) error {
	replayAttackParams, ok := params.(*ReplayAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *ReplayAttackParams")
	}
	return replayAttackParams.Validate()
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

func (p *StoreAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("invalid message code: %d", p.Code)
	}

	return nil
}

func (p *StoreAttackParams) ValidateWith(params interface{}) error {
	storeAttackParams, ok := params.(*StoreAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *StoreAttackParams")
	}
	return storeAttackParams.Validate()
}

// DosAttackParams represents parameters for DoS attack
type DosAttackParams struct {
	Code  MessageCode `json:"code"`
	Cmd   uint64      `json:"cmd"`   // 0(valid), 1(sequence change), 2(round change)
	Cnt   uint64      `json:"cnt"`   // number of messages to send
	Delay uint64      `json:"delay"` // delay in milliseconds
}

var _ AttackParamsParser = (*DosAttackParams)(nil)

// HasMessageCode checks if the given message code matches
func (p *DosAttackParams) HasMessageCode(code MessageCode) bool {
	return p.Code.Has(code)
}

func (p *DosAttackParams) UnmarshalJSON(data []byte) error {
	type Alias DosAttackParams
	return json.Unmarshal(data, (*Alias)(p))
}

// Parse parses raw parameters into DosAttackParams
func (p *DosAttackParams) Parse(raw map[string]interface{}) (interface{}, error) {
	params := &DosAttackParams{}
	params.Code = ParseMessageCode(raw["code"])

	// Parse cmd
	if cmd, err := ParseToUint64(raw["cmd"]); err == nil {
		params.Cmd = cmd
	} else {
		return nil, fmt.Errorf("invalid cmd value: %w", err)
	}

	// Parse cnt
	if cnt, err := ParseToUint64(raw["cnt"]); err == nil {
		params.Cnt = cnt
	} else {
		return nil, fmt.Errorf("invalid cnt value: %w", err)
	}

	// Parse delay
	if delay, err := ParseToUint64(raw["delay"]); err == nil {
		params.Delay = delay
	} else {
		return nil, fmt.Errorf("invalid delay value: %w", err)
	}

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
func (p *DosAttackParams) Validate() error {
	if !ValidateMessageCode(p.Code) {
		return fmt.Errorf("dosParam: invalid message code: %d", p.Code)
	}

	// Validate cmd
	if p.Cmd < 0 || p.Cmd > 2 {
		return fmt.Errorf("dosParam: invalid cmd: %d, must be 0, 1, or 2", p.Cmd)
	}

	if p.Cnt == 0 {
		return fmt.Errorf("dosParam: cnt must be greater than 0")
	}

	return nil
}

// ValidateWith validates DosAttackParams against another instance
func (p *DosAttackParams) ValidateWith(params interface{}) error {
	dosAttackParams, ok := params.(*DosAttackParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: expected *DosAttackParams")
	}
	return dosAttackParams.Validate()
}

// Helper function

func ParseToUint64(rawValue interface{}) (uint64, error) {
	switch v := rawValue.(type) {
	case float64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint64:
		return v, nil
	case json.Number:
		val, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("cannot parse json.Number to int64: %w", err)
		}
		return uint64(val), nil
	default:
		return 0, fmt.Errorf("unsupported type for uint64 conversion: %T", v)
	}
}

func ParseTargets(rawTargets interface{}) []string {
	var targets []string

	switch v := rawTargets.(type) {
	case []interface{}:
		for _, target := range v {
			if str, ok := target.(string); ok {
				// Accept both addresses and enode URLs
				targets = append(targets, str)
			}
		}
	case []string:
		// Directly use string array (can be addresses or enode URLs)
		targets = v
	case []common.Address:
		// Convert addresses to strings
		for _, addr := range v {
			targets = append(targets, addr.Hex())
		}
	}

	return targets
}

// ParseFakeBLSKeys parses the value field to extract or generate fake BLS keys
// Returns a slice of FakeBLSKey containing address, secret key, and public key
func ParseFakeBLSKeys(value interface{}) ([]*FakeBLSKey, bool) {
	// Check if value is nil
	if value == nil {
		// Generate random BLS key
		secretKey, err := bls.GenerateKey(nil)
		if err != nil {
			return nil, false
		}
		publicKey := secretKey.PublicKey()
		addr := utils.GenerateRandomAddress()
		return []*FakeBLSKey{{
			Address:   addr,
			SecretKey: secretKey,
			PublicKey: publicKey,
		}}, true
	}

	switch v := value.(type) {
	case string:
		// Check if it's "nil" string
		if v == "nil" {
			key, _ := crypto.GenerateKey()
			// Generate random BLS key
			blsKey, err := bls.DeriveFromECDSA(key)
			if err != nil {
				return nil, false
			}
			publicKey := blsKey.PublicKey()
			addr := utils.GenerateRandomAddress()
			return []*FakeBLSKey{{
				Address:   addr,
				SecretKey: blsKey,
				PublicKey: publicKey,
			}}, true
		}
		// Try to parse as hex BLS secret key (longer than address)
		if strings.HasPrefix(v, "0x") && len(v) > 42 {
			// Remove 0x prefix
			hexStr := strings.TrimPrefix(v, "0x")
			if keyBytes, err := hex.DecodeString(hexStr); err == nil && len(keyBytes) == 32 {
				if secretKey, err := bls.GenerateKey(keyBytes); err == nil {
					publicKey := secretKey.PublicKey()
					// Generate address from public key or use provided one
					addr := utils.GenerateRandomAddress()
					return []*FakeBLSKey{{
						Address:   addr,
						SecretKey: secretKey,
						PublicKey: publicKey,
					}}, true
				}
			}
		}
		// Otherwise treat as address and generate BLS key
		if common.IsHexAddress(v) {
			addr := common.HexToAddress(v)
			// Generate deterministic BLS key from address
			seed := make([]byte, 32)
			copy(seed[:20], addr.Bytes())
			secretKey, err := bls.GenerateKey(seed)
			if err != nil {
				return nil, false
			}
			publicKey := secretKey.PublicKey()
			return []*FakeBLSKey{{
				Address:   addr,
				SecretKey: secretKey,
				PublicKey: publicKey,
			}}, true
		}
	case []string:
		// Multiple BLS keys or addresses
		keys := make([]*FakeBLSKey, 0, len(v))
		for _, item := range v {
			if item == "nil" {
				// Generate random BLS key for "nil" entry
				secretKey, err := bls.GenerateKey(nil)
				if err != nil {
					continue
				}
				publicKey := secretKey.PublicKey()
				addr := utils.GenerateRandomAddress()
				keys = append(keys, &FakeBLSKey{
					Address:   addr,
					SecretKey: secretKey,
					PublicKey: publicKey,
				})
			} else if strings.HasPrefix(item, "0x") && len(item) > 42 {
				// Try to parse as BLS secret key
				hexStr := strings.TrimPrefix(item, "0x")
				if keyBytes, err := hex.DecodeString(hexStr); err == nil && len(keyBytes) == 32 {
					if secretKey, err := bls.GenerateKey(keyBytes); err == nil {
						publicKey := secretKey.PublicKey()
						addr := utils.GenerateRandomAddress()
						keys = append(keys, &FakeBLSKey{
							Address:   addr,
							SecretKey: secretKey,
							PublicKey: publicKey,
						})
					}
				}
			} else if common.IsHexAddress(item) {
				// Generate BLS key from address
				addr := common.HexToAddress(item)
				seed := make([]byte, 32)
				copy(seed[:20], addr.Bytes())
				secretKey, err := bls.GenerateKey(seed)
				if err != nil {
					continue
				}
				publicKey := secretKey.PublicKey()
				keys = append(keys, &FakeBLSKey{
					Address:   addr,
					SecretKey: secretKey,
					PublicKey: publicKey,
				})
			}
		}
		if len(keys) > 0 {
			return keys, true
		}
	case []interface{}:
		// Handle []interface{} case
		keys := make([]*FakeBLSKey, 0, len(v))
		for _, item := range v {
			switch val := item.(type) {
			case string:
				if val == "nil" {
					// Generate random BLS key for "nil" entry
					secretKey, err := bls.GenerateKey(nil)
					if err != nil {
						continue
					}
					publicKey := secretKey.PublicKey()
					addr := utils.GenerateRandomAddress()
					keys = append(keys, &FakeBLSKey{
						Address:   addr,
						SecretKey: secretKey,
						PublicKey: publicKey,
					})
				} else if strings.HasPrefix(val, "0x") && len(val) > 42 {
					// Try to parse as BLS secret key
					hexStr := strings.TrimPrefix(val, "0x")
					if keyBytes, err := hex.DecodeString(hexStr); err == nil && len(keyBytes) == 32 {
						if secretKey, err := bls.GenerateKey(keyBytes); err == nil {
							publicKey := secretKey.PublicKey()
							addr := utils.GenerateRandomAddress()
							keys = append(keys, &FakeBLSKey{
								Address:   addr,
								SecretKey: secretKey,
								PublicKey: publicKey,
							})
						}
					}
				} else if common.IsHexAddress(val) {
					// Generate BLS key from address
					addr := common.HexToAddress(val)
					seed := make([]byte, 32)
					copy(seed[:20], addr.Bytes())
					secretKey, err := bls.GenerateKey(seed)
					if err != nil {
						continue
					}
					publicKey := secretKey.PublicKey()
					keys = append(keys, &FakeBLSKey{
						Address:   addr,
						SecretKey: secretKey,
						PublicKey: publicKey,
					})
				}
			case map[string]interface{}:
				// Handle object format with address and secretKey fields
				if addrStr, ok := val["address"].(string); ok {
					addr := common.HexToAddress(addrStr)
					if secretKeyStr, ok := val["secretKey"].(string); ok {
						// Parse provided secret key
						hexStr := strings.TrimPrefix(secretKeyStr, "0x")
						if keyBytes, err := hex.DecodeString(hexStr); err == nil && len(keyBytes) == 32 {
							if secretKey, err := bls.GenerateKey(keyBytes); err == nil {
								publicKey := secretKey.PublicKey()
								keys = append(keys, &FakeBLSKey{
									Address:   addr,
									SecretKey: secretKey,
									PublicKey: publicKey,
								})
							}
						}
					} else {
						// Generate BLS key from address
						seed := make([]byte, 32)
						copy(seed[:20], addr.Bytes())
						secretKey, err := bls.GenerateKey(seed)
						if err != nil {
							continue
						}
						publicKey := secretKey.PublicKey()
						keys = append(keys, &FakeBLSKey{
							Address:   addr,
							SecretKey: secretKey,
							PublicKey: publicKey,
						})
					}
				}
			}
		}
		if len(keys) > 0 {
			return keys, true
		}
	}
	return nil, false
}

func validateTargets(targets interface{}) error {
	switch v := targets.(type) {
	case []string:
		for i, target := range v {
			if target == "" {
				return fmt.Errorf("invalid target at index %d: empty string", i)
			}
			// Validate if it's an address or enode URL
			if !common.IsHexAddress(target) && !strings.HasPrefix(target, "enode://") {
				return fmt.Errorf("invalid target at index %d: must be address or enode URL", i)
			}
		}
	case []common.Address:
		for i, target := range v {
			if target == (common.Address{}) {
				return fmt.Errorf("invalid target address at index %d: address %v", i, target.Hex())
			}
		}
	default:
		return fmt.Errorf("invalid targets type: %T", targets)
	}
	return nil
}
