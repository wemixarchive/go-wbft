package api

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
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
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	// Convert fields
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
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	// Convert fields
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
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	// Convert fields
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
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	return params, nil
}

func ConvertToRoleSpoofParams(raw map[string]interface{}) (types.RoleSpoofParams, error) {
	params := types.RoleSpoofParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	// Convert fields
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

	return params, nil
}

func ConvertToReplayMessageParams(raw map[string]interface{}) (types.ReplayMessageParams, error) {
	params := types.ReplayMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	return params, nil
}

func ConvertToStoreMessageParams(raw map[string]interface{}) (types.StoreMessageParams, error) {
	params := types.StoreMessageParams{}

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
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

	// Set enabled (default to true if not specified)
	if v, exists := raw["enabled"]; exists {
		if enabled, ok := v.(bool); ok {
			params.Enabled = enabled
		} else {
			params.Enabled = false
		}
	} else {
		return params, fmt.Errorf("required field \"enabled\"")
	}

	if v, exists := raw["seq_s"]; exists {
		seqStart, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_s: %w", err)
		}
		params.SequenceStart = seqStart
	} else {
		return params, fmt.Errorf("seq_s is required")
	}

	if v, exists := raw["seq_e"]; exists {
		seqEnd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid seq_e: %w", err)
		}
		params.SequenceEnd = seqEnd
	} else {
		return params, fmt.Errorf("seq_e is required")
	}

	// Parse round
	if v, ok := raw["round"].(uint64); ok {
		params.Round = v
	} else {
		return params, fmt.Errorf("round number is required")
	}

	// Parse code
	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("invalid code")
	}

	// Parse cmd
	if v, exists := raw["cmd"]; exists {
		cmd, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid cmd: %w", err)
		}
		params.Cmd = cmd
	} else {
		params.Cmd = 0 // default to 0
	}

	// Parse cnt (count)
	if v, exists := raw["cnt"]; exists {
		cnt, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid cnt: %w", err)
		}
		params.Cnt = cnt
	} else {
		params.Cnt = 1 // default to 1
	}

	// Parse delay
	if v, exists := raw["delay"]; exists {
		delay, err := convertToUint64(v)
		if err != nil {
			return params, fmt.Errorf("invalid delay: %w", err)
		}
		params.Delay = delay
	} else {
		params.Delay = 0 // default to 0
	}

	return params, nil
}
