package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// TestPrepareWarmCoinbase checks that the coinbase starts warm when Croissant
// or Shanghai (EIP-3651) is active, while the sender is always warm.
func TestPrepareWarmCoinbase(t *testing.T) {
	sender := common.HexToAddress("0x1111111111111111111111111111111111111111")
	coinbase := common.HexToAddress("0x2222222222222222222222222222222222222222")

	for _, tt := range []struct {
		name         string
		rules        params.Rules
		coinbaseWarm bool
	}{
		{"berlin", params.Rules{IsBerlin: true}, false},
		{"croissant", params.Rules{IsBerlin: true, IsCroissant: true}, true},
		{"shanghai", params.Rules{IsBerlin: true, IsShanghai: true}, true},
	} {
		statedb, err := New(types.EmptyRootHash, NewDatabase(rawdb.NewMemoryDatabase()), nil)
		if err != nil {
			t.Fatal(err)
		}
		statedb.Prepare(tt.rules, sender, coinbase, nil, nil, nil)

		if !statedb.AddressInAccessList(sender) {
			t.Errorf("%s: sender must be warm", tt.name)
		}
		if got := statedb.AddressInAccessList(coinbase); got != tt.coinbaseWarm {
			t.Errorf("%s: coinbase warm = %v, want %v", tt.name, got, tt.coinbaseWarm)
		}
	}
}
