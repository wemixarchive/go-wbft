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
// This file is derived from quorum/consensus/istanbul/config.go (2024.07.25).
// Modified and improved for the wemix development.

package qbft

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"

	"github.com/naoina/toml"
)

type ProposerPolicyId uint64

const (
	RoundRobin ProposerPolicyId = iota
	Sticky
)

// ProposerPolicy represents the Validator Proposer Policy
type ProposerPolicy struct {
	Id ProposerPolicyId    // Could be RoundRobin or Sticky
	By ValidatorSortByFunc // func that defines how the ValidatorSet should be sorted
}

// NewRoundRobinProposerPolicy returns a RoundRobin ProposerPolicy with ValidatorSortByString as default sort function
func NewRoundRobinProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(RoundRobin)
}

// NewStickyProposerPolicy return a Sticky ProposerPolicy with ValidatorSortByString as default sort function
func NewStickyProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(Sticky)
}

func NewProposerPolicy(id ProposerPolicyId) *ProposerPolicy {
	return NewProposerPolicyByIdAndSortFunc(id, ValidatorSortByString())
}

func NewProposerPolicyByIdAndSortFunc(id ProposerPolicyId, by ValidatorSortByFunc) *ProposerPolicy {
	return &ProposerPolicy{Id: id, By: by}
}

type proposerPolicyToml struct {
	Id ProposerPolicyId
}

func (p *ProposerPolicy) MarshalTOML() (interface{}, error) {
	if p == nil {
		return nil, nil
	}
	pp := &proposerPolicyToml{Id: p.Id}
	data, err := toml.Marshal(pp)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (p *ProposerPolicy) UnmarshalTOML(decode func(interface{}) error) error {
	var innerToml string
	err := decode(&innerToml)
	if err != nil {
		return err
	}
	var pp proposerPolicyToml
	err = toml.Unmarshal([]byte(innerToml), &pp)
	if err != nil {
		return err
	}
	p.Id = pp.Id
	p.By = ValidatorSortByString()
	return nil
}

// Use sets the ValidatorSortByFunc for the given ProposerPolicy and sorts the validatorSets according to it
func (p *ProposerPolicy) Use(v ValidatorSortByFunc) {
	p.By = v
}

type Config struct {
	RequestTimeout           uint64                  `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod              uint64                  `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	ProposerPolicy           *ProposerPolicy         `toml:",omitempty"` // The policy for proposer selection
	Epoch                    uint64                  `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
	AllowedFutureBlockTime   uint64                  `toml:",omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BlockReward              *math.HexOrDecimal256   `toml:",omitempty"` // Reward
	BlockRewardBeneficiary   *params.BeneficiaryInfo `toml:",omitempty"`
	TargetValidators         uint64                  `toml:",omitempty"`
	MaxRequestTimeoutSeconds uint64                  `toml:",omitempty"`
	UseNCP                   bool                    `toml:",omitempty"` // Use NCP or not
	Transitions              []params.Transition
}

var DefaultConfig = &Config{
	RequestTimeout:         1000,
	BlockPeriod:            1,
	ProposerPolicy:         NewRoundRobinProposerPolicy(),
	Epoch:                  10,
	AllowedFutureBlockTime: 0,
}

func (c Config) GetConfig(blockNumber *big.Int) Config {
	newConfig := c

	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.RequestTimeoutSeconds != 0 {
			// RequestTimeout is on milliseconds
			newConfig.RequestTimeout = transition.RequestTimeoutSeconds * 1000
		}
		if transition.EpochLength != 0 {
			newConfig.Epoch = transition.EpochLength
		}
		if transition.BlockPeriodSeconds != 0 {
			newConfig.BlockPeriod = transition.BlockPeriodSeconds
		}
		if transition.BlockReward != nil {
			newConfig.BlockReward = transition.BlockReward
		}
		if transition.BlockRewardBeneficiary != nil {
			newConfig.BlockRewardBeneficiary = transition.BlockRewardBeneficiary
		}
		if transition.TargetValidators != nil {
			newConfig.TargetValidators = *transition.TargetValidators
		}
		if transition.MaxRequestTimeoutSeconds != nil {
			newConfig.MaxRequestTimeoutSeconds = *transition.MaxRequestTimeoutSeconds
		}
		newConfig.UseNCP = transition.UseNCP
	})

	return newConfig
}

func (c *Config) getTransitionValue(num *big.Int, callback func(transition params.Transition)) {
	if c != nil && num != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(num) <= 0; i++ {
			callback(c.Transitions[i])
		}
	}
}

// String implements the stringer interface, returning the consensus engine details.
func (c *Config) String() string {
	return "wbft"
}

func GetMontBlancTransition(chainConfig *params.ChainConfig, num *big.Int) (*params.StateTransition, error) {
	if chainConfig == nil || chainConfig.MontBlancBlock == nil || num == nil {
		return nil, errors.New("nil montBlanc config or nil block number")
	}

	if num.Cmp(chainConfig.MontBlancBlock) == 0 {
		return getMontBlancTransition(chainConfig.MontBlanc.Init.GovContracts)
	}

	for _, upgrade := range chainConfig.MontBlanc.Upgrades {
		if num.Cmp(upgrade.Block) == 0 {
			return getMontBlancTransition(upgrade.GovContracts)
		} else if num.Cmp(upgrade.Block) < 0 {
			break
		}
	}
	return nil, nil
}

func getMontBlancTransition(govContracts *params.GovContracts) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if govContracts.GovConfig != nil {
		minStaking, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_MINIMUM_STAKING], 10)
		maxStaking, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_MAXIMUM_STAKING], 10)
		unbondingStaker, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_UNBONDING_STAKER], 10)
		unbondingDelegator, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_UNBONDING_DELEGATOR], 10)
		feePrecision, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_FEE_PRECISION], 10)
		changeFeeDelay, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_CHANGE_FEE_DELAY], 10)
		stabilizingStakerThreshold, _ := new(big.Int).SetString(govContracts.GovConfig.Params[gov.GOV_CONFIG_PARAM_STABILIZING_STAKER_THRESHOLD], 10)
		if minStaking == nil || maxStaking == nil || unbondingStaker == nil || unbondingDelegator == nil ||
			feePrecision == nil || changeFeeDelay == nil || stabilizingStakerThreshold == nil {
			return nil, errors.New("invalid gov config params")
		}

		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovConfig.Address, Code: govwbft.GovContractCodes[gov.CONTRACT_GOV_CONFIG][govContracts.GovConfig.Version]})
		st.States = append(st.States, []params.StateParam{
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(0)), Value: common.BigToHash(minStaking)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(1)), Value: common.BigToHash(maxStaking)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(2)), Value: common.BigToHash(unbondingStaker)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(3)), Value: common.BigToHash(unbondingDelegator)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(4)), Value: common.BigToHash(feePrecision)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(5)), Value: common.BigToHash(changeFeeDelay)},
			{Address: govContracts.GovConfig.Address, Key: common.BigToHash(big.NewInt(6)), Value: common.BigToHash(stabilizingStakerThreshold)},
		}...)
	}

	if govContracts.GovStaking != nil {
		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovStaking.Address, Code: govwbft.GovContractCodes[gov.CONTRACT_GOV_STAKING][govContracts.GovStaking.Version]})

		// initialize govConfig, govRewardeeImp addresses of GovStaking contract
		if govContracts.GovConfig != nil {
			st.States = append(st.States, params.StateParam{
				Address: govContracts.GovStaking.Address, Key: common.HexToHash(govwbft.SLOT_GOV_CONFIG_ADDRESS), Value: common.BytesToHash(govContracts.GovConfig.Address.Bytes())})
		}
		if govContracts.GovRewardeeImp != nil {
			st.States = append(st.States, params.StateParam{
				Address: govContracts.GovStaking.Address, Key: common.HexToHash(govwbft.SLOT_GOV_REWARDEE_IMP_ADDRESS), Value: common.BytesToHash(govContracts.GovRewardeeImp.Address.Bytes())})
		}
	}

	if govContracts.GovRewardeeImp != nil {
		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovRewardeeImp.Address, Code: govwbft.GovContractCodes[gov.CONTRACT_GOV_REWARDEE_IMP][govContracts.GovRewardeeImp.Version]})
	}

	if govContracts.GovNCP != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: govContracts.GovNCP.Address, Code: govwbft.GovContractCodes[gov.CONTRACT_GOV_NCP][govContracts.GovNCP.Version]})
		ncpAddresses := strings.Split(govContracts.GovNCP.Params[gov.GOV_NCP_PARAM_NCPS], ",")
		ncps := make([]common.Address, 0)
		for _, ncp := range ncpAddresses {
			ncps = append(ncps, common.HexToAddress(ncp))
		}
		st.States = append(st.States, govwbft.InitializeNCP(govContracts.GovNCP.Address, ncps)...)
	}
	return st, nil
}
