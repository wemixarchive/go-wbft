package params

import (
	"math/big"
	"testing"
)

// londonConfig returns a chain config with every fork up to london at genesis,
// so CheckConfigForkOrder only judges the WEMIX forks a test adds on top.
func londonConfig() *ChainConfig {
	zero := big.NewInt(0)
	return &ChainConfig{
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

// TestIsCroissantGating checks that Croissant is gated by block number, unlike
// the timestamp-gated Shanghai fork on Ethereum mainnet.
func TestIsCroissantGating(t *testing.T) {
	const forkBlock = 100
	cfg := &ChainConfig{CroissantBlock: big.NewInt(forkBlock)}
	for _, tt := range []struct {
		num  int64
		want bool
	}{
		{0, false},
		{forkBlock - 1, false},
		{forkBlock, true},
		{forkBlock + 1, true},
	} {
		if got := cfg.IsCroissant(big.NewInt(tt.num)); got != tt.want {
			t.Errorf("IsCroissant(%d) = %v, want %v", tt.num, got, tt.want)
		}
	}

	unset := &ChainConfig{}
	if unset.IsCroissant(big.NewInt(1 << 40)) {
		t.Error("IsCroissant must be false when CroissantBlock is nil")
	}
}

// TestCroissantRulesPropagation checks that Rules carries the Croissant flag,
// which is what the EVM and the state transition actually read. The flag does
// not depend on the merge or on the block time.
func TestCroissantRulesPropagation(t *testing.T) {
	const forkBlock = 10
	cfg := londonConfig()
	cfg.CroissantBlock = big.NewInt(forkBlock)
	if cfg.Rules(big.NewInt(forkBlock-1), false, 0).IsCroissant {
		t.Error("Rules.IsCroissant must be false before the fork block")
	}
	if !cfg.Rules(big.NewInt(forkBlock), false, 0).IsCroissant {
		t.Error("Rules.IsCroissant must be true at the fork block")
	}
}

// TestCroissantCompatibility checks that rescheduling an already passed
// Croissant block is reported as an incompatible config change.
func TestCroissantCompatibility(t *testing.T) {
	stored := londonConfig()
	stored.CroissantBlock = big.NewInt(10)
	moved := londonConfig()
	moved.CroissantBlock = big.NewInt(20)

	if err := stored.CheckCompatible(moved, 5, 0); err != nil {
		t.Errorf("moving a fork that has not been reached must be allowed: %v", err)
	}
	if err := stored.CheckCompatible(moved, 15, 0); err == nil {
		t.Error("moving a fork the head has already passed must be rejected")
	}
}

// TestWemixNetworksScheduleCroissant checks that mainnet and testnet both
// declare a Croissant block before Dalgona and pass the fork order check.
func TestWemixNetworksScheduleCroissant(t *testing.T) {
	for name, cfg := range map[string]*ChainConfig{
		"mainnet": WemixMainnetChainConfig,
		"testnet": WemixTestnetChainConfig,
	} {
		if cfg.CroissantBlock == nil {
			t.Errorf("%s: CroissantBlock is nil", name)
		}
		if err := cfg.CheckConfigForkOrder(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestCroissantBeforeDalgona checks that a config scheduling Dalgona must also
// schedule Croissant at or before it.
func TestCroissantBeforeDalgona(t *testing.T) {
	for _, tt := range []struct {
		name      string
		croissant *big.Int
		dalgona   *big.Int
		wantErr   bool
	}{
		{"neither", nil, nil, false},
		{"croissant only", big.NewInt(10), nil, false},
		{"croissant before dalgona", big.NewInt(10), big.NewInt(20), false},
		{"same block", big.NewInt(10), big.NewInt(10), false},
		{"dalgona without croissant", nil, big.NewInt(20), true},
		{"croissant after dalgona", big.NewInt(30), big.NewInt(20), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := londonConfig()
			cfg.CroissantBlock = tt.croissant
			cfg.DalgonaBlock = tt.dalgona
			err := cfg.CheckConfigForkOrder()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckConfigForkOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
