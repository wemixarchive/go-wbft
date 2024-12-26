package backend

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

// GetValidators Retrieve the Validator List from the Extra field of the EpochBlock's Header.
func (sb *Backend) GetValidators(chain consensus.ChainHeaderReader, blockNumber *big.Int, hash common.Hash) (qbft.ValidatorSet, error) {
	// 1. Return an empty address set if the (Montblanc) HardFork is not supported
	if !chain.Config().IsMontBlanc(blockNumber) && chain.Config().MontBlancBlock != nil {
		emptyValSet := make([]common.Address, 0)
		return validator.NewSet(emptyValSet, sb.config.ProposerPolicy), nil
	}

	// 2. Retrieve the QBFT configuration for a specific block number
	qbftConfig := sb.config.GetConfig(blockNumber)

	// 3. Calculate the ValidatorSet based on the current state
	var valSet qbft.ValidatorSet
	{
		// 3-1. Retrieve the nearest EpochBlock for the given block number.
		//      : (n)th EpochBlock == Last Block of the (n-1)th Epoch
		//      : (n)th EpochBlock == Start Block of the (n)th Epoch - 1
		nearestEpochBlock, err := qbftConfig.GetNearestEpochBlock(chain, blockNumber.Uint64())
		if err != nil {
			log.Error("BFT: not found epochBlock", "err", err)
			return nil, err
		}

		// Return the QBFT Config from chainConfig if the nearest epochBlock is not detected in transitions.
		if nearestEpochBlock.Sign() == 0 {
			return validator.NewSet(qbftConfig.Validators, sb.config.ProposerPolicy), nil
		}

		// 3-2. Retrieve the header of the nearest EpochBlock
		epochHeader := chain.GetHeaderByNumber(nearestEpochBlock.Uint64())
		if epochHeader == nil {
			log.Error("BFT: not found header", "blocknumber", nearestEpochBlock.Uint64())
			return nil, errors.New("BFT: not found header")
		}

		// 3-3. Extract the Extra field from the Header to obtain the ValidatorSet
		if qbftExtra, err := types.ExtractQBFTExtra(epochHeader); err == nil {
			valSet = validator.NewSet(qbftExtra.Validators, qbftConfig.ProposerPolicy)
		} else {
			log.Error("BFT: invalid epoch header", "err", err)
			return nil, err
		}
	}
	return valSet, nil
}
