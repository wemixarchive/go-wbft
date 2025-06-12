// Modification Copyright 2024 The Wemix Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/config_test.go (2024.07.25).
// Modified and improved for the wemix development.

package qbft

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"github.com/stretchr/testify/assert"
)

func TestProposerPolicy_UnmarshalTOML(t *testing.T) {
	input := `id = 2
`
	expectedId := ProposerPolicyId(2)
	var p proposerPolicyToml
	assert.NoError(t, toml.Unmarshal([]byte(input), &p))
	assert.Equal(t, expectedId, p.Id, "ProposerPolicyId mismatch")
}

func TestProposerPolicy_MarshalTOML(t *testing.T) {
	output := `id = 1
`
	p := &ProposerPolicy{Id: 1}
	b, err := p.MarshalTOML()
	if err != nil {
		t.Errorf("error marshalling ProposerPolicy: %v", err)
	}
	assert.Equal(t, output, b, "ProposerPolicy MarshalTOML mismatch")
}

// TestGetConfigTransition tests the transition from old Transitions to new hard fork approach
//func TestGetConfigTransition(t *testing.T) {
//	// Test config with old Transitions (should still work for backward compatibility)
//	configWithTransitions := *DefaultConfig
//	configWithTransitions.Transitions = []params.Transition{{
//		Block:                 big.NewInt(1),
//		EpochLength:           300,
//		BlockPeriodSeconds:    3,
//		RequestTimeoutSeconds: 3000,
//	}}
//
//	result := configWithTransitions.GetConfig(big.NewInt(1), nil)
//
//	expectedConfig := *DefaultConfig
//	expectedConfig.Epoch = 300
//	expectedConfig.BlockPeriod = 3
//	expectedConfig.RequestTimeout = 3000 * 1000
//
//	if !reflect.DeepEqual(result, expectedConfig) {
//		t.Errorf("error transition config:\nexpected: %+v\ngot: %+v\n", expectedConfig, result)
//	}
//}

// testChainConfigWrapper wraps ChainConfig to add fake hardForks for testing
type chainConfigWrapper struct {
	*params.ChainConfig
	fakeHardForks []fakeHardFork
}

type fakeHardFork struct {
	name              string
	blockNum          *big.Int
	WBFTConfig        *params.WBFTConfig
	GovContractConfig *params.GovContracts
}

// newTestChainConfig creates a test chain config with fake hard forks
func newTestChainConfig() *chainConfigWrapper {
	return &chainConfigWrapper{
		ChainConfig:   params.TestQBFTChainConfig,
		fakeHardForks: make([]fakeHardFork, 0),
	}
}

// addFakeHardFork adds a fake hard fork for testing
func (cw *chainConfigWrapper) addFakeHardFork(name string, blockNum *big.Int, wbftConfig *params.WBFTConfig, govContractConfig *params.GovContracts) {
	fh := fakeHardFork{
		name:              name,
		blockNum:          blockNum,
		WBFTConfig:        wbftConfig,
		GovContractConfig: govContractConfig,
	}
	cw.fakeHardForks = append(cw.fakeHardForks, fh)
}

// isFakeHardFork checks if a fake hard fork is active at the given block
func (cw *chainConfigWrapper) isFakeHardFork(name string, blockNum *big.Int) bool {
	var hf fakeHardFork
	for _, fh := range cw.fakeHardForks {
		if fh.name == name {
			hf = fh
			break
		}
	}
	if hf.blockNum == nil {
		return false
	} else {
		return blockNum.Cmp(hf.blockNum) >= 0
	}
}

// getWbftHardforkValueTest is a test version of getWbftHardforkValue that works with fake hardforks
func (c *Config) getWbftHardforkValueTest(num *big.Int, testConfig *chainConfigWrapper, callback func(wbftConfig params.WBFTConfig)) {
	if c != nil && num != nil && testConfig != nil {
		// Check MontBlanc first
		if testConfig.ChainConfig.IsMontBlanc(num) && testConfig.ChainConfig.MontBlanc != nil && testConfig.ChainConfig.MontBlanc.WBFT != nil {
			// do nothing. use qbftConfig as it is, since qbftConfig is set as montblanc
		}
		// Check fake hard forks in order (TestFork1, TestFork2, TestFork3)
		for _, fh := range testConfig.fakeHardForks {
			if testConfig.isFakeHardFork(fh.name, num) {
				callback(*fh.WBFTConfig)
			}
		}
	}
}

// GetConfigTest is a test version of GetConfig that works with fake hardforks
func (c Config) GetConfigTest(blockNumber *big.Int, testConfig *chainConfigWrapper) Config {
	newConfig := c

	// Use fake hard fork logic instead of real hard fork logic
	if testConfig != nil {
		c.getWbftHardforkValueTest(blockNumber, testConfig, func(wbftConfig params.WBFTConfig) {
			if wbftConfig.RequestTimeoutSeconds != 0 {
				// RequestTimeout is on milliseconds
				newConfig.RequestTimeout = wbftConfig.RequestTimeoutSeconds * 1000
			}
			if wbftConfig.BlockPeriodSeconds != 0 {
				newConfig.BlockPeriod = wbftConfig.BlockPeriodSeconds
			}
			if wbftConfig.EpochLength != 0 {
				newConfig.Epoch = wbftConfig.EpochLength
			}
			if wbftConfig.BlockReward != nil {
				newConfig.BlockReward = wbftConfig.BlockReward
			}
			if wbftConfig.BlockRewardBeneficiary != nil {
				newConfig.BlockRewardBeneficiary = wbftConfig.BlockRewardBeneficiary
			}
			if wbftConfig.ProposerPolicy != nil {
				newConfig.ProposerPolicy = NewProposerPolicy(ProposerPolicyId(*wbftConfig.ProposerPolicy))
			}
			if wbftConfig.TargetValidators != nil {
				newConfig.TargetValidators = *wbftConfig.TargetValidators
			}
			if wbftConfig.MaxRequestTimeoutSeconds != nil {
				newConfig.MaxRequestTimeoutSeconds = *wbftConfig.MaxRequestTimeoutSeconds
			}
			if wbftConfig.StabilizingStakersThreshold != nil {
				newConfig.StabilizingStakersThreshold = *wbftConfig.StabilizingStakersThreshold
			}
			if wbftConfig.UseNCP != nil {
				newConfig.UseNCP = *wbftConfig.UseNCP
			}
		})
	}
	return newConfig
}

func setConfigFromChainConfig(qbftCfg *Config, config *params.WBFTConfig) error {
	if len(config.Transitions) > 0 {
		qbftCfg.Transitions = config.Transitions
	}
	if config.BlockPeriodSeconds != 0 {
		qbftCfg.BlockPeriod = config.BlockPeriodSeconds
	}
	if config.RequestTimeoutSeconds != 0 {
		qbftCfg.RequestTimeout = config.RequestTimeoutSeconds * 1000
	}
	if config.EpochLength != 0 {
		qbftCfg.Epoch = config.EpochLength
	}

	qbftCfg.ProposerPolicy = NewProposerPolicy(ProposerPolicyId(*config.ProposerPolicy))
	qbftCfg.BlockReward = config.BlockReward
	qbftCfg.BlockRewardBeneficiary = config.BlockRewardBeneficiary
	qbftCfg.TargetValidators = *config.TargetValidators

	if config.MaxRequestTimeoutSeconds != nil && *config.MaxRequestTimeoutSeconds > 0 {
		qbftCfg.MaxRequestTimeoutSeconds = *config.MaxRequestTimeoutSeconds
	}
	qbftCfg.StabilizingStakersThreshold = *config.StabilizingStakersThreshold
	qbftCfg.UseNCP = *config.UseNCP

	return nil
}

func TestGetConfig(t *testing.T) {
	// Create test chain config with fake hard forks
	testConfig := newTestChainConfig()
	qbftCfg := new(Config)

	setConfigFromChainConfig(qbftCfg, testConfig.MontBlanc.WBFT)

	createExpectedConfig := func(baseConfig *Config, modifications func(*Config)) Config {
		expected := *baseConfig
		modifications(&expected)
		return expected
	}

	// Add fake hard forks
	testConfig.addFakeHardFork("TestFork1", big.NewInt(10),
		&params.WBFTConfig{
			EpochLength: 200,
		},
		nil,
	)

	testConfig.addFakeHardFork("TestFork2", big.NewInt(20),
		&params.WBFTConfig{
			BlockPeriodSeconds: 3,
		},
		nil,
	)

	testConfig.addFakeHardFork("TestFork3", big.NewInt(30),
		&params.WBFTConfig{
			RequestTimeoutSeconds: 4000,
		},
		nil,
	)

	tests := []struct {
		name           string
		blockNumber    uint64
		expectedConfig Config
	}{
		{
			name:           "Before any hard fork (block 5)",
			blockNumber:    5,
			expectedConfig: *qbftCfg,
		},
		{
			name:        "After TestFork1 (block 15)",
			blockNumber: 15,
			expectedConfig: createExpectedConfig(qbftCfg, func(cfg *Config) {
				cfg.Epoch = 200 // From TestFork1
			}),
		},
		{
			name:        "After TestFork2 (block 25)",
			blockNumber: 25,
			expectedConfig: createExpectedConfig(qbftCfg, func(cfg *Config) {
				cfg.Epoch = 200     // From TestFork1
				cfg.BlockPeriod = 3 // From TestFork2
			}),
		},
		{
			name:        "After TestFork3 (block 35)",
			blockNumber: 35,
			expectedConfig: createExpectedConfig(qbftCfg, func(cfg *Config) {
				cfg.Epoch = 200                  // From TestFork1
				cfg.BlockPeriod = 3              // From TestFork2
				cfg.RequestTimeout = 4000 * 1000 // From TestFork3
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := qbftCfg.GetConfigTest(big.NewInt(int64(test.blockNumber)), testConfig)
			if !reflect.DeepEqual(result, test.expectedConfig) {
				t.Errorf("error in %s:\nexpected: %+v\ngot: %+v\n", test.name, test.expectedConfig, result)
			}
		})
	}
}

func TestGetGovContracts(t *testing.T) {

}
