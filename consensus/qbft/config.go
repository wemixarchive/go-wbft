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
	"github.com/ethereum/go-ethereum/consensus"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
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
	RequestTimeout           uint64                `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod              uint64                `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	EmptyBlockPeriod         uint64                `toml:",omitempty"` // Default minimum difference between a block and empty block's timestamps in second
	ProposerPolicy           *ProposerPolicy       `toml:",omitempty"` // The policy for proposer selection
	Epoch                    uint64                `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
	AllowedFutureBlockTime   uint64                `toml:",omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BeneficiaryMode          *string               `toml:",omitempty"` // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	BlockReward              *math.HexOrDecimal256 `toml:",omitempty"` // Reward
	MiningBeneficiary        *common.Address       `toml:",omitempty"` // Wallet address that benefits at every new block (besu mode)
	Validators               []common.Address      `toml:",omitempty"`
	ValidatorSelectionMode   *string               `toml:",omitempty"`
	Client                   bind.ContractCaller   `toml:",omitempty"`
	MaxRequestTimeoutSeconds uint64                `toml:",omitempty"`
	Transitions              []params.Transition
}

var DefaultConfig = &Config{
	RequestTimeout:         10000,
	BlockPeriod:            5,
	EmptyBlockPeriod:       0,
	ProposerPolicy:         NewRoundRobinProposerPolicy(),
	Epoch:                  30000,
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
		if transition.EmptyBlockPeriodSeconds != nil {
			newConfig.EmptyBlockPeriod = *transition.EmptyBlockPeriodSeconds
		}
		if transition.BeneficiaryMode != nil {
			newConfig.BeneficiaryMode = transition.BeneficiaryMode
		}
		if transition.BlockReward != nil {
			newConfig.BlockReward = transition.BlockReward
		}
		if transition.MiningBeneficiary != nil {
			newConfig.MiningBeneficiary = transition.MiningBeneficiary
		}
		if transition.ValidatorSelectionMode != "" {
			newConfig.ValidatorSelectionMode = &transition.ValidatorSelectionMode
		}
		if len(transition.Validators) > 0 {
			newConfig.Validators = transition.Validators
		}
		if transition.MaxRequestTimeoutSeconds != nil {
			newConfig.MaxRequestTimeoutSeconds = *transition.MaxRequestTimeoutSeconds
		}
	})

	return newConfig
}

func (c Config) GetValidatorSelectionMode(blockNumber *big.Int) string {
	mode := params.BlockHeaderMode
	if c.ValidatorSelectionMode != nil {
		mode = *c.ValidatorSelectionMode
	}
	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.ValidatorSelectionMode != "" {
			mode = transition.ValidatorSelectionMode
		}
	})
	return mode
}

func (c Config) GetValidatorsAt(blockNumber *big.Int) []common.Address {
	if blockNumber.Cmp(big.NewInt(0)) == 0 && len(c.Validators) > 0 {
		return c.Validators
	}

	if blockNumber != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(blockNumber) == 0; i++ {
			return c.Transitions[i].Validators
		}
	}

	//Note! empty means we will get the valset from previous block header which contains votes, validators etc
	return []common.Address{}
}

func (c *Config) getTransitionValue(num *big.Int, callback func(transition params.Transition)) {
	if c != nil && num != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(num) <= 0; i++ {
			callback(c.Transitions[i])
		}
	}
}

// IsEpochBlock checks whether the provided block number corresponds to an EpochBlock.
//
// Defined the EpochBlock as the block that records the ValidatorList for the Nth Epoch.
// Specifies that the EpochBlock for the Nth Epoch is the last block of the (N-1)th Epoch.
//
// Returns true if the block number corresponds to an EpochBlock, otherwise false.
func (c *Config) IsEpochBlock(chain consensus.ChainHeaderReader, blockNumber *big.Int) bool {

	// 1. Retrieves the starting block number of the hard fork (associated with the given block number)
	startForkBlock := c.GetNearestForkBlock(chain, blockNumber)
	if startForkBlock.Sign() == 0 {
		return true
	}

	// 2. Calculate the offset relative to the fork start block.
	offset := new(big.Int).Sub(blockNumber, startForkBlock)
	if offset.Sign() == 0 { // starting block of the HardFork
		return true
	}

	if offset.Sign() == -1 {
		return false
	}

	// 3. Check if block itself is an EpochBlock
	epochInterval := new(big.Int).SetUint64(c.Epoch)
	return new(big.Int).Mod(offset, epochInterval).Cmp(epochInterval) == -1
}

// GetNearestForkBlock retrieves the starting block number of the hard fork closest to the provided block number.
//
// Returns:
//   - This function guarantees that it will never return a nil value.
func (c *Config) GetNearestForkBlock(chain consensus.ChainHeaderReader, blockNumber *big.Int) *big.Int {

	nearestForkBlock := c.getNearestForkBlock(blockNumber)

	// If the associated HardFork of the block cannot be found, retrieve it from the chainConfig
	if nearestForkBlock == nil {
		switch {
		case chain.Config().IsMontBlanc(blockNumber):
			return chain.Config().MontBlancBlock
		default:
			return big.NewInt(0)
		}
	}
	return nearestForkBlock
}

// getNearestForkBlock is the internal implementation of GetNearestForkBlock.
// assumes that the epoch value is only changed in a HardFork block, and not in any other block.
//
// Returns:
// - It can be nil if the nearest hard fork block number corresponding to the given block number does not exist.
func (c *Config) getNearestForkBlock(blockNumber *big.Int) *big.Int {
	// TODO: (noah) transition의 Block 순서를 보장하지 못하는 경우를 가정하여 만듬
	var nearestForkBlock *big.Int = nil

	for _, transition := range c.Transitions {
		if transition.Block.Cmp(blockNumber) == 0 {
			return transition.Block
		}
		if transition.Block.Cmp(blockNumber) < -1 && transition.Block.Cmp(nearestForkBlock) > 0 {
			nearestForkBlock = transition.Block
		}
	}
	return nearestForkBlock
}

// GetNearestEpochBlock returns the nearest EpochBlock (for the given block number)
//
// Defined the EpochBlock as the block that records the ValidatorList for the Nth Epoch.
// Specifies that the EpochBlock for the Nth Epoch is the last block of the (N-1)th Epoch.
//
// Parameters:
//   - block : Block number for which the nearest EpochBlock is to be found (for the (N)th Epoch).
//
// Returns:
//   - epochBlock : Block number of the last block in the (N-1)th Epoch, which serves as the EpochBlock for the (N)th Epoch.
func (c *Config) GetNearestEpochBlock(chain consensus.ChainHeaderReader, block uint64) (*big.Int, error) {

	if block == 0 {
		return big.NewInt(0), nil
	}
	blockNumber := new(big.Int).SetUint64(block)

	// 1. Retrieves the starting block number of the hard fork (associated with the given block number)
	startForkBlock := c.GetNearestForkBlock(chain, blockNumber)
	if startForkBlock.Sign() == 0 {
		return startForkBlock, nil
	}

	if blockNumber.Cmp(startForkBlock) < 0 {
		return nil, errors.New("invalid blockNumber: The BlockNumber must exceed the starting BlockNumber of its HardFork")
	}

	// 2. Calculate the offset from the fork start block
	offset := new(big.Int).Sub(blockNumber, startForkBlock)
	if offset.Sign() == 0 { // starting block of the HardFork
		return new(big.Int).SetUint64(block - 1), nil
	}

	// 3. Check if block itself is an EpochBlock
	epochInterval := new(big.Int).SetUint64(c.Epoch)
	if new(big.Int).Mod(offset, epochInterval).Sign() == 0 { // starting block of a certain Epoch
		return new(big.Int).SetUint64(block - 1), nil
	}

	// 4. Calculate the nearest EpochBlock (based on the given offset and epoch interval)
	//  = startForkBlock + ( (offset / epoch) * epoch )
	q := new(big.Int).Div(offset, epochInterval)
	nearestEpochBlock := new(big.Int).Add(startForkBlock, q.Mul(q, epochInterval))

	return nearestEpochBlock.Sub(nearestEpochBlock, big.NewInt(1)), nil
}

// String implements the stringer interface, returning the consensus engine details.
func (c *Config) String() string {
	return "qbft"
}

func GetStateTransitions(chainConfig *params.ChainConfig, num *big.Int) []params.StateTransition {
	if chainConfig != nil && num != nil {
		transitions := make([]params.StateTransition, 0)

		if chainConfig.MontBlancBlock != nil && chainConfig.MontBlancBlock.Cmp(num) == 0 {
			transitions = append(transitions, getMontBlancTransition(chainConfig.MontBlanc))
		}

		if st := chainConfig.GetStateTransitions(num); len(st) > 0 {
			transitions = append(transitions, st...)
		}
		return transitions
	}
	return nil
}

func getMontBlancTransition(config *params.MontBlancConfig) params.StateTransition {
	st := params.StateTransition{
		Codes: []params.CodeParam{
			{Address: govwbft.GovConstAddress, Code: govwbft.GovConstContract},
			{Address: govwbft.GovStakingAddress, Code: govwbft.GovStakingContract},
		},
	}
	if config != nil && len(config.NCPs) > 0 {
		st.Codes = append(st.Codes, params.CodeParam{Address: govwbft.GovNCPAddress, Code: govwbft.GovNCPContract})
		st.States = govwbft.InitializeNCP(config.NCPs)
	}
	return st
}
