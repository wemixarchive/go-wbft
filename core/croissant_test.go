package core

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// croissantForkBlock is where the test chain activates Croissant.
const croissantForkBlock = 10

var (
	croissantTestSender   = common.HexToAddress("0x1000000000000000000000000000000000000001")
	croissantTestCoinbase = common.HexToAddress("0x2000000000000000000000000000000000000002")
)

// croissantChainConfig is a london-era chain config, like WEMIX mainnet before
// Dalgona, that activates Croissant at croissantForkBlock.
func croissantChainConfig() *params.ChainConfig {
	zero := big.NewInt(0)
	return &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      zero,
		EIP150Block:         zero,
		EIP155Block:         zero,
		EIP158Block:         zero,
		ByzantiumBlock:      zero,
		ConstantinopleBlock: zero,
		PetersburgBlock:     zero,
		IstanbulBlock:       zero,
		BerlinBlock:         zero,
		LondonBlock:         zero,
		CroissantBlock:      big.NewInt(croissantForkBlock),
	}
}

// TestIntrinsicGasEIP3860 checks that the per-word initcode charge applies to
// contract creation when EIP-3860 is on, and never to ordinary calls.
func TestIntrinsicGasEIP3860(t *testing.T) {
	const words = 2
	data := make([]byte, words*32)
	wantDelta := words * params.InitCodeWordGas

	for _, tt := range []struct {
		name     string
		creation bool
		delta    uint64
	}{
		{"creation", true, wantDelta},
		{"call", false, 0},
	} {
		pre, err := IntrinsicGas(data, nil, nil, tt.creation, true, true, false)
		if err != nil {
			t.Fatalf("%s without EIP-3860: %v", tt.name, err)
		}
		post, err := IntrinsicGas(data, nil, nil, tt.creation, true, true, true)
		if err != nil {
			t.Fatalf("%s with EIP-3860: %v", tt.name, err)
		}
		if post-pre != tt.delta {
			t.Errorf("%s: EIP-3860 adds %d gas, want %d", tt.name, post-pre, tt.delta)
		}
	}
}

// applyAt runs a message from croissantTestSender at block number on
// croissantChainConfig. A nil to means contract creation. code, when given, is
// installed at to before the message runs.
func applyAt(t *testing.T, block int64, to *common.Address, code, data []byte) (*ExecutionResult, error) {
	t.Helper()
	statedb, err := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		t.Fatal(err)
	}
	if to != nil && code != nil {
		statedb.SetCode(*to, code)
	}
	const gasLimit = 30_000_000
	blockCtx := vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		Coinbase:    croissantTestCoinbase,
		BlockNumber: big.NewInt(block),
		Difficulty:  big.NewInt(1), // go-wemix PoA headers carry difficulty 1
		GasLimit:    gasLimit,
		BaseFee:     big.NewInt(0),
	}
	evm := vm.NewEVM(blockCtx, vm.TxContext{GasPrice: big.NewInt(0)}, statedb, croissantChainConfig(), vm.Config{NoBaseFee: true})

	zero := big.NewInt(0)
	msg := &Message{
		From:      croissantTestSender,
		To:        to,
		Value:     zero,
		GasLimit:  gasLimit,
		GasPrice:  zero,
		GasFeeCap: zero,
		GasTipCap: zero,
		Data:      data,
	}
	return ApplyMessage(evm, msg, new(GasPool).AddGas(math.MaxUint64))
}

// TestEIP3860CreationTxRejected checks that a creation transaction with
// initcode over the limit is invalid from the fork on: ApplyMessage returns a
// consensus error instead of including it as a failed execution.
func TestEIP3860CreationTxRejected(t *testing.T) {
	oversized := make([]byte, params.MaxInitCodeSize+1)

	if _, err := applyAt(t, croissantForkBlock, nil, nil, oversized); !errors.Is(err, ErrMaxInitCodeSizeExceeded) {
		t.Fatalf("oversized initcode after the fork: got %v, want ErrMaxInitCodeSizeExceeded", err)
	}
	for _, tt := range []struct {
		name     string
		block    int64
		initcode []byte
	}{
		{"oversized before the fork", croissantForkBlock - 1, oversized},
		{"at the limit after the fork", croissantForkBlock, make([]byte, params.MaxInitCodeSize)},
	} {
		res, err := applyAt(t, tt.block, nil, nil, tt.initcode)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if res.Failed() {
			t.Fatalf("%s: execution failed: %v", tt.name, res.Err)
		}
	}
}
