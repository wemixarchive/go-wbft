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
// This file is derived from quorum/core/types/istanbul_test.go (2024.07.25).
// Modified and improved for the wemix development.

package types

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// ## Quorum QBFT START
func TestHeaderHash(t *testing.T) {
	// 0x06f58ebb96f233516c4a0eeedad007e8b8e0965b24c88d7407eb08a6ee7bc4ff
	expectedExtra := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000f89af8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b440b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0")
	expectedHash := common.HexToHash("0x06f58ebb96f233516c4a0eeedad007e8b8e0965b24c88d7407eb08a6ee7bc4ff")

	// for istanbul consensus
	header := &Header{Difficulty: QBFTDefaultDifficulty, Extra: expectedExtra}
	if !reflect.DeepEqual(header.Hash(), expectedHash) {
		t.Errorf("expected: %v, but got: %v", expectedHash.Hex(), header.Hash().Hex())
	}

	// append useless information to extra-data
	unexpectedExtra := append(expectedExtra, []byte{1, 2, 3}...)
	header.Extra = unexpectedExtra
	if !reflect.DeepEqual(header.Hash(), rlpHash(header)) {
		t.Errorf("expected: %v, but got: %v", rlpHash(header).Hex(), header.Hash().Hex())
	}
}

func Genesis(validators []common.Address) *core.Genesis {
	// generate genesis block
	genesis := core.DefaultGenesisBlock()
	genesis.Config = params.TestChainConfig
	// force enable QBFT engine
	genesis.Config.QBFT = &params.QBFTConfig{}
	genesis.Config.Ethash = nil
	genesis.Difficulty = QBFTDefaultDifficulty
	genesis.Nonce = BlockNonce{}.Uint64()

	appendValidators(genesis, validators)

	return genesis
}

func appendValidators(genesis *core.Genesis, addrs []common.Address) {
	vanity := append(genesis.ExtraData, bytes.Repeat([]byte{0x00}, IstanbulExtraVanity-len(genesis.ExtraData))...)
	ist := &QBFTExtra{
		VanityData:    vanity,
		Validators:    addrs,
		Rewards:       make([]common.Address, 0),
		Vote:          nil,
		CommittedSeal: [][]byte{},
		Round:         0,
	}

	istPayload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		panic("failed to encode qbft extra")
	}
	genesis.ExtraData = istPayload
}

func TestExtractToQBFTExtra(t *testing.T) {
	testCases := []struct {
		istRawData     []byte
		expectedResult *QBFTExtra
		expectedErr    error
	}{
		{
			// normal case
			hexutil.MustDecode("0xf89b80f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b440f83f9444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6c080c0"),
			&QBFTExtra{
				VanityData: []byte{},
				Validators: []common.Address{
					common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
					common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
					common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
					common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440")),
				},
				Rewards: []common.Address{
					common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
					common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
					common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
				},
				CommittedSeal: [][]byte{},
				Round:         0,
				Vote:          nil,
			},
			nil,
		},
		{
			// normal case
			hexutil.MustDecode("0xf89b80f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b440f83f9444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6c080c0"),
			&QBFTExtra{
				VanityData: []byte{},
				Validators: []common.Address{
					common.BytesToAddress(hexutil.MustDecode("0x76ec8c6b805bc28eb62e9ec276073c4cdf717603")),
					common.BytesToAddress(hexutil.MustDecode("0xd5c40819e657cd7178b182b6bfc529c81e708fb1")),
					common.BytesToAddress(hexutil.MustDecode("0x79a1f1510b845f60fa85a7dbe2bc6ad9896fd726")),
					common.BytesToAddress(hexutil.MustDecode("0x35365eab5af106b3edaed5e2e221a3391e8b6309")),
				},
				Rewards: []common.Address{
					common.BytesToAddress(hexutil.MustDecode("0x76ec8c6b805bc28eb62e9ec276073c4cdf717603")),
					common.BytesToAddress(hexutil.MustDecode("0xd5c40819e657cd7178b182b6bfc529c81e708fb1")),
					common.BytesToAddress(hexutil.MustDecode("0x79a1f1510b845f60fa85a7dbe2bc6ad9896fd726")),
				},
				CommittedSeal: [][]byte{},
				Round:         0,
				Vote:          nil,
			},
			nil,
		},
	}
	for _, test := range testCases {
		//tt, _ := rlp.EncodeToBytes(test.expectedResult)
		//fmt.Println(hex.EncodeToString(tt))
		fmt.Println(Genesis(test.expectedResult.Validators).ExtraData)
		h := &Header{Extra: test.istRawData}
		istanbulExtra, err := ExtractQBFTExtra(h)
		if err != test.expectedErr {
			t.Errorf("expected: %v, but got: %v", test.expectedErr, err)
		}
		if !reflect.DeepEqual(istanbulExtra, test.expectedResult) {
			t.Errorf("expected: %v, but got: %v", test.expectedResult, istanbulExtra)
		}
	}
}

// ## Quorum QBFT END
