package core

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
)

func GetValidatorsFromState(state govwbft.StateReader) []common.Address {
	return govwbft.NCPStakers(state)
}

// VerifyValidators checks whether the ValidatorList matches the ValidatorList in the state.
func VerifyValidators(validators []common.Address, state govwbft.StateReader) error {
	validatorFromState := GetValidatorsFromState(state)

	sort := func(addrs []common.Address) {
		for i := 0; i < len(addrs); i++ {
			for j := i + 1; j < len(addrs); j++ {
				if bytes.Compare(addrs[i][:], addrs[j][:]) > 0 {
					addrs[i], addrs[j] = addrs[j], addrs[i]
				}
			}
		}
	}
	// Checks if two arrays have the same elements in the same order
	{
		// Check if the lengths are different
		if len(validators) != len(validatorFromState) {
			return errors.New("WBFT: mismatch in ValidatorList sizes")
		}

		sort(validatorFromState)
		sort(validators)

		// Compare each element
		for i := range validators {
			if validators[i] != validatorFromState[i] {
				return errors.New("WBFT: The two validators do not match")
			}
		}
	}
	return nil
}
