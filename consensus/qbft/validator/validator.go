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
// This file is derived from quorum/consensus/istanbul/validator/validator.go (2024.07.25).
// Modified and improved for the wemix development.

package validator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
)

func New(addr common.Address, blsPublicKey []byte) qbft.Validator {
	return &defaultValidator{
		address:      addr,
		blsPublicKey: blsPublicKey,
	}
}

//func New(addr common.Address, blsPublicKey []string) qbft.Validator {
//	return &defaultValidator{
//		address:      addr,
//		blsPublicKey: blsPublicKey,
//	}
//}

func NewSet(addrs []common.Address, blsPublicKeys [][]byte, policy *qbft.ProposerPolicy) qbft.ValidatorSet {
	validators := make(qbft.Validators, len(addrs))
	for i, addr := range addrs {
		validators[i] = New(addr, blsPublicKeys[i])
	}
	return newDefaultSet(validators, policy)
}

//func NewSet(addrs []common.Address, blsPublicKeys []string, policy *qbft.ProposerPolicy) qbft.ValidatorSet {
//	if len(addrs) != len(blsPublicKeys) {
//		log.Error("mismatched slice lengths: addrs=%d, blsPublicKeys=%d", len(addrs), len(blsPublicKeys))
//	}
//	if policy == nil {
//		log.Error("proposer policy must not be nil")
//	}
//
//	validators := make(qbft.Validators, len(addrs))
//
//	for i, addr := range addrs {
//		pkBytes, _ := hexutil.Decode(blsPublicKeys[i])
//		validators[i] = New(addr, pkBytes)
//	}
//
//	return newDefaultSet(validators, policy)
//}

func NewSetByValidators(validators qbft.Validators, policy *qbft.ProposerPolicy) qbft.ValidatorSet {
	return newDefaultSet(validators, policy)
}
