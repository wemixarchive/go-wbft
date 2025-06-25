package api

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// ConvertToSilentMessageParams converts raw map to typed params
func ConvertToSilentMessageParams(raw map[string]interface{}) (types.SilentMessageParams, error) {
	params := types.SilentMessageParams{}

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
	} else {
		return params, fmt.Errorf("invalid code")
	}

	if v, ok := raw["direction"].(float64); ok {
		params.Direction = uint64(v)
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

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
	} else {
		return params, fmt.Errorf("invalid code")
	}

	// Convert tamper fields
	if v, ok := raw["tamperFields"].([]interface{}); ok {
		params.TamperFields = make([]types.TamperField, len(v))
		for i, field := range v {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				if target, ok := fieldMap["target"].(string); ok {
					params.TamperFields[i].Target = target
				}
				params.TamperFields[i].Value = fieldMap["value"]
			} else {
				return params, fmt.Errorf("invalid tamper field at index %d", i)
			}
		}
	}

	if v, ok := raw["withValidMessage"].(bool); ok {
		params.WithValidMessage = v
	}

	if v, ok := raw["delay"].(float64); ok {
		params.Delay = uint64(v)
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

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
	} else {
		return params, fmt.Errorf("invalid code")
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

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
	} else {
		return params, fmt.Errorf("invalid code")
	}

	if v, ok := raw["cmd"].(float64); ok {
		params.Cmd = uint64(v)
	} else {
		return params, fmt.Errorf("invalid cmd")
	}

	if v, ok := raw["cnt"].(float64); ok {
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

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
	} else {
		return params, fmt.Errorf("invalid code")
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

	if v, ok := raw["oriSequence"].(float64); ok {
		params.OriSequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid oriSequence")
	}

	if v, ok := raw["oriRound"].(float64); ok {
		params.OriRound = uint64(v)
	} else {
		return params, fmt.Errorf("invalid oriRound")
	}

	if v, ok := raw["sequence"].(float64); ok {
		params.Sequence = uint64(v)
	} else {
		return params, fmt.Errorf("invalid sequence")
	}

	if v, ok := raw["round"].(float64); ok {
		params.Round = uint64(v)
	} else {
		return params, fmt.Errorf("invalid round")
	}

	if v, ok := raw["useOriginalView"].(bool); ok {
		params.UseOriginalView = v
	}

	if v, ok := raw["code"].(float64); ok {
		params.Code = uint64(v)
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
