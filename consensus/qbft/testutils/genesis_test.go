package testutils

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestGeneratingGenesisExtra(t *testing.T) {
	QBFTExtra := &types.QBFTExtra{
		VanityData: []byte{},
		Validators: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0xa62987A40E094CbE020313f19F71aeeB3E48B86f")),
			common.BytesToAddress(hexutil.MustDecode("0xE1bD8108149FCa703B9FCD8ea967B6b7e660f13b")),
			common.BytesToAddress(hexutil.MustDecode("0xd19f9374f4549B2fB182ED766d6b7501494a3634")),
			common.BytesToAddress(hexutil.MustDecode("0x02cF1E577C79EF0E93947cCd82a4D41E0485Be73")),
		},
		Rewards: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0xa62987A40E094CbE020313f19F71aeeB3E48B86f")),
			common.BytesToAddress(hexutil.MustDecode("0xE1bD8108149FCa703B9FCD8ea967B6b7e660f13b")),
			common.BytesToAddress(hexutil.MustDecode("0xd19f9374f4549B2fB182ED766d6b7501494a3634")),
		},
		CommittedSeal: [][]byte{},
		Round:         0,
		Vote:          nil,
	}
	genesis := GensissWithRewards(QBFTExtra.Validators, QBFTExtra.Rewards)
	//fmt.Println(hex.EncodeToString(genesis.ExtraData))
	qbftExtra := new(types.QBFTExtra)
	err := rlp.DecodeBytes(genesis.ExtraData[:], qbftExtra)
	if err != nil {
		t.Errorf("Failed to decode genesis qbft extra data : %v", err)
	} else if len(qbftExtra.Rewards) != len(QBFTExtra.Rewards) {
		t.Errorf("QBFT Extra is not properly endcoded")
	}
}
