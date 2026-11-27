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

// TestIsCroissantFromDalgona checks that Dalgona carries the Croissant EIPs, so
// IsCroissant holds from the Dalgona block on even when Croissant is unset or
// scheduled later.
func TestIsCroissantFromDalgona(t *testing.T) {
	const dalgona = 100
	for _, tt := range []struct {
		name      string
		croissant *big.Int
		num       int64
		want      bool
	}{
		{"croissant unset, before dalgona", nil, dalgona - 1, false},
		{"croissant unset, at dalgona", nil, dalgona, true},
		{"croissant after dalgona, at dalgona", big.NewInt(dalgona + 50), dalgona, true},
		{"croissant before dalgona, between them", big.NewInt(dalgona - 50), dalgona - 1, true},
	} {
		cfg := &ChainConfig{CroissantBlock: tt.croissant, DalgonaBlock: big.NewInt(dalgona)}
		if got := cfg.IsCroissant(big.NewInt(tt.num)); got != tt.want {
			t.Errorf("%s: IsCroissant(%d) = %v, want %v", tt.name, tt.num, got, tt.want)
		}
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

// TestWemixNetworksScheduleCroissant pins the Croissant block of mainnet and
// testnet, keeps the Dalgona block that selects the WEMIX consensus engine, and
// checks the fork order. go-wemix must use the same Croissant blocks.
func TestWemixNetworksScheduleCroissant(t *testing.T) {
	for _, tt := range []struct {
		name      string
		cfg       *ChainConfig
		croissant int64
	}{
		{"mainnet", WemixMainnetChainConfig, 134_004_000},
		{"testnet", WemixTestnetChainConfig, 138_903_600},
	} {
		if tt.cfg.CroissantBlock == nil || tt.cfg.CroissantBlock.Int64() != tt.croissant {
			t.Errorf("%s: CroissantBlock = %v, want %d", tt.name, tt.cfg.CroissantBlock, tt.croissant)
		}
		if !tt.cfg.DalgonaEnabled() {
			t.Errorf("%s: Dalgona must stay configured to select the WEMIX consensus engine", tt.name)
		}
		if err := tt.cfg.CheckConfigForkOrder(); err != nil {
			t.Errorf("%s: %v", tt.name, err)
		}
	}
}

// TestCroissantOrderNotEnforced checks that CheckConfigForkOrder accepts
// Croissant before, at or after Dalgona, or unset, because Dalgona carries the
// Croissant EIPs itself.
func TestCroissantOrderNotEnforced(t *testing.T) {
	for _, tt := range []struct {
		name      string
		croissant *big.Int
		dalgona   *big.Int
	}{
		{"neither", nil, nil},
		{"croissant only", big.NewInt(10), nil},
		{"croissant before dalgona", big.NewInt(10), big.NewInt(20)},
		{"same block", big.NewInt(10), big.NewInt(10)},
		{"dalgona without croissant", nil, big.NewInt(20)},
		{"croissant after dalgona", big.NewInt(30), big.NewInt(20)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := londonConfig()
			cfg.CroissantBlock = tt.croissant
			cfg.DalgonaBlock = tt.dalgona
			if err := cfg.CheckConfigForkOrder(); err != nil {
				t.Errorf("CheckConfigForkOrder() = %v, want nil", err)
			}
		})
	}
}

// TestMaxInitCodeSizeDerivation pins EIP-3860's MAX_INITCODE_SIZE to twice
// WEMIX's MaxCodeSize, the same value go-wemix uses. A different value would
// split the chain on a CREATE whose initcode falls between the two limits.
func TestMaxInitCodeSizeDerivation(t *testing.T) {
	const wemixMaxInitCodeSize = 507904
	if MaxInitCodeSize != 2*MaxCodeSize {
		t.Fatalf("MaxInitCodeSize = %d, want 2*MaxCodeSize = %d", MaxInitCodeSize, 2*MaxCodeSize)
	}
	if MaxInitCodeSize != wemixMaxInitCodeSize {
		t.Fatalf("MaxInitCodeSize = %d, want %d", MaxInitCodeSize, wemixMaxInitCodeSize)
	}
}
