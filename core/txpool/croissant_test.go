package txpool

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// TestValidateTransactionInitCodeLimit checks that the pool rejects a creation
// transaction whose initcode exceeds params.MaxInitCodeSize once the head is at
// Croissant, or at Dalgona when Croissant is unset, and accepts it before.
// MaxSize is raised above the initcode limit so the size check does not reject
// the transaction first.
func TestValidateTransactionInitCodeLimit(t *testing.T) {
	const forkBlock = 10
	for _, tt := range []struct {
		name      string
		croissant *big.Int
		dalgona   *big.Int
	}{
		{"croissant", big.NewInt(forkBlock), nil},
		{"dalgona without croissant", nil, big.NewInt(forkBlock)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := londonConfig()
			config.CroissantBlock = tt.croissant
			config.DalgonaBlock = tt.dalgona
			key, err := crypto.GenerateKey()
			if err != nil {
				t.Fatal(err)
			}
			signer := types.LatestSigner(config)
			tx, err := types.SignNewTx(key, signer, &types.LegacyTx{
				Gas:      10_000_000,
				GasPrice: big.NewInt(params.InitialBaseFee),
				Data:     make([]byte, params.MaxInitCodeSize+1),
			})
			if err != nil {
				t.Fatal(err)
			}
			opts := &ValidationOptions{
				Config:  config,
				Accept:  1 << types.LegacyTxType,
				MaxSize: 2 * params.MaxInitCodeSize,
				MinTip:  new(big.Int),
			}
			head := func(num int64) *types.Header {
				return &types.Header{Number: big.NewInt(num), GasLimit: 30_000_000}
			}

			if err := ValidateTransaction(tx, head(forkBlock-1), signer, opts); err != nil {
				t.Fatalf("before the fork: %v", err)
			}
			if err := ValidateTransaction(tx, head(forkBlock), signer, opts); !errors.Is(err, core.ErrMaxInitCodeSizeExceeded) {
				t.Fatalf("at the fork: got %v, want ErrMaxInitCodeSizeExceeded", err)
			}
		})
	}
}

// londonConfig returns a chain config with every fork up to london at genesis.
func londonConfig() *params.ChainConfig {
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
	}
}
