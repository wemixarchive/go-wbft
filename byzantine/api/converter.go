package api

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// convertToUint64 converts various numeric types to uint64
func convertToUint64(v interface{}) (uint64, error) {
	switch val := v.(type) {
	case uint64:
		return val, nil
	case float64:
		return uint64(val), nil
	case int:
		return uint64(val), nil
	case int64:
		return uint64(val), nil
	case uint:
		return uint64(val), nil
	case uint32:
		return uint64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", v)
	}
}

// ConvertToSetMessagePolicyParams converts raw map to typed params
func ConvertToSetMessagePolicyParams(raw map[string]interface{}) (types.SetMessagePolicyParams, error) {
	params := types.SetMessagePolicyParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, exists := raw["code"]; exists {
		code, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid code: %w", err)
		}
		params.Code = code
	} else {
		return params, fmt.Errorf("code is required")
	}

	// Convert fields (supports both 'fields' and 'tamperFields' for backward compatibility)
	var fields []interface{}
	if v, ok := raw["fields"].([]interface{}); ok {
		fields = v
	}

	if fields != nil {
		params.Fields = make([]types.Field, 0, len(fields))

		for i, field := range fields {
			fieldMap, ok := field.(map[string]interface{})
			if !ok {
				return params, fmt.Errorf("invalid field at index %d", i)
			}

			targetStr, ok := fieldMap["target"].(string)
			if !ok {
				return params, fmt.Errorf("missing or invalid target at index %d", i)
			}

			f := types.Field{
				Target: targetStr,
				Value:  fieldMap["value"],
			}
			params.Fields = append(params.Fields, f)
		}
	} else {
		return params, fmt.Errorf("fields is required")
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}
	return params, nil
}

// ConvertToTamperedMessageParams converts raw map to typed params
func ConvertToTamperedMessageParams(raw map[string]interface{}) (types.TamperedMessageParams, error) {
	params := types.TamperedMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	// Support both single sequence and range
	if v, exists := raw["sequence"]; exists {
		seq, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid sequence: %w", err)
		}
		params.Sequence = seq
	} else if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart

		if v, exists := raw["seq_e"]; exists {
			seqEnd, err := convertToUint64(v)
			if err != nil {
				return params, fmt.Errorf("invalid seq_e: %w", err)
			}
			params.SequenceEnd = seqEnd
		}
	} else {
		return params, fmt.Errorf("sequence or seq_s is required")
	}

	if v, exists := raw["round"]; exists {
		round, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid round: %w", err)
		}
		params.Round = round
	}

	if v, exists := raw["code"]; exists {
		code, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid code: %w", err)
		}
		params.Code = code
	} else {
		return params, fmt.Errorf("code is required")
	}

	// Convert fields (supports both 'fields' and 'tamperFields' for backward compatibility)
	var fields []interface{}
	if v, ok := raw["fields"].([]interface{}); ok {
		fields = v
	} else if v, ok := raw["tamperFields"].([]interface{}); ok {
		fields = v
	}

	if fields != nil {
		params.Fields = make([]types.Field, 0, len(fields))

		for i, field := range fields {
			fieldMap, ok := field.(map[string]interface{})
			if !ok {
				return params, fmt.Errorf("invalid field at index %d", i)
			}

			targetStr, ok := fieldMap["target"].(string)
			if !ok {
				return params, fmt.Errorf("missing or invalid target at index %d", i)
			}

			f := types.Field{
				Target: targetStr,
				Value:  fieldMap["value"],
			}
			params.Fields = append(params.Fields, f)
		}
	}
	if v, ok := raw["withValidMessage"].(bool); ok {
		params.WithValidMessage = v
	}

	if v, exists := raw["delay"]; exists {
		delay, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid delay: %w", err)
		}
		params.Delay = delay
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}

	return params, nil
}

// ConvertToFakeMessageParams converts raw map to typed params
func ConvertToFakeMessageParams(raw map[string]interface{}) (types.FakeMessageParams, error) {
	params := types.FakeMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("sequence number is required")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("round number is required")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("code number is required")
	}

	if v, ok := raw["fakeMessage"]; ok {
		// Handle both string and byte array
		switch msg := v.(type) {
		case string:
			params.FakeMessage = []byte(msg)
		case []byte:
			params.FakeMessage = msg
		default:
			return params, fmt.Errorf("invalid fakeMessage type: %T", v)
		}
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}

	return params, nil
}

// Additional converter functions for other param types...
// ConvertToOmitMessageParams, ConvertToRoleSpoofParams, ConvertToReplayMessageParams

func ConvertToOmitMessageParams(raw map[string]interface{}) (types.OmitMessageParams, error) {
	params := types.OmitMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("sequence number is required")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("round number is required")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("code number is required")
	}

	if v, ok := raw["cmd"].(uint64); ok {
		params.Cmd = v
	} else {
		return params, fmt.Errorf("cmd is required")
	}

	if v, ok := raw["cnt"].(uint64); ok {
		params.Cnt = uint64(v)
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}

	return params, nil
}

func ConvertToRoleSpoofParams(raw map[string]interface{}) (types.RoleSpoofParams, error) {
	params := types.RoleSpoofParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("sequence number is required")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("round number is required")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("code number is required")
	}

	if v, ok := raw["fakeMessage"]; ok {
		switch msg := v.(type) {
		case string:
			params.FakeMessage = []byte(msg)
		case []byte:
			params.FakeMessage = msg
		default:
			return params, fmt.Errorf("invalid fakeMessage type: %T", v)
		}
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}

	return params, nil
}

func ConvertToReplayMessageParams(raw map[string]interface{}) (types.ReplayMessageParams, error) {
	params := types.ReplayMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("invalid code")
	}

	if v, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = v
	}

	// Convert targets
	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			}
		}
	}

	return params, nil
}

func ConvertToStoreMessageParams(raw map[string]interface{}) (types.StoreMessageParams, error) {
	params := types.StoreMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["sequence"].(uint64); ok {
		params.Sequence = v
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("invalid code")
	}
	return params, nil
}

// ConvertToDosMessageParams converts API parameters to DosAttackParams
func ConvertToDosMessageParams(raw map[string]interface{}) (types.DosMessageParams, error) {
	params := types.DosMessageParams{}

	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = true
		}
	} else {
		params.Enabled = true
	}

	if v, ok := raw["seq_s"].(uint64); ok {
		params.SequenceStart = v
	} else {
		return params, fmt.Errorf("invalid seq_s")
	}

	if v, ok := raw["seq_e"].(uint64); ok {
		params.SequenceEnd = v
	} else {
		return params, fmt.Errorf("invalid seq_e")
	}

	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("round number is required")
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("invalid code")
	}

	// Parse fields
	if fieldsValue, ok := raw["fields"].([]interface{}); ok {
		params.Fields = make([]types.Field, 0, len(fieldsValue))
		for _, f := range fieldsValue {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				field := types.Field{}
				if target, ok := fieldMap["target"].(string); ok {
					field.Target = target
				}
				if value, ok := fieldMap["value"]; ok {
					field.Value = value
				}
				params.Fields = append(params.Fields, field)
			}
		}
	}

	return params, nil
}
