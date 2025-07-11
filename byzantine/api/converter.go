package api

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// ConvertToSilentMessageParams converts raw map to typed params
func ConvertToSilentMessageParams(raw map[string]interface{}) (types.SilentMessageParams, error) {
	params := types.SilentMessageParams{}

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

	if v, ok := raw["direction"].(uint64); ok {
		params.Direction = v
	} else {
		return params, fmt.Errorf("invalid direction")
	}

	if v, ok := raw["targets"].([]interface{}); ok {
		params.Targets = make([]common.Address, len(v))
		for i, addr := range v {
			if strAddr, ok := addr.(string); ok {
				params.Targets[i] = common.HexToAddress(strAddr)
			} else {
				return params, fmt.Errorf("invalid target address at index %d", i)
			}
		}
	}

	return params, nil
}

// ConvertToTamperedMessageParams converts raw map to typed params
func ConvertToTamperedMessageParams(raw map[string]interface{}) (types.TamperedMessageParams, error) {
	params := types.TamperedMessageParams{}

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

	if v, ok := raw["delay"].(uint64); ok {
		params.Delay = v
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

	if v, ok := raw["oriSequence"].(uint64); ok {
		params.OriSequence = v
	} else {
		return params, fmt.Errorf("invalid oriSequence")
	}

	if v, ok := raw["oriRound"].(uint64); ok {
		params.OriRound = v
	} else {
		return params, fmt.Errorf("invalid oriRound")
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

	if v, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = v
	}

	if v, ok := raw["code"].(uint64); ok {
		params.Code = v
	} else {
		return params, fmt.Errorf("invalid code")
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
