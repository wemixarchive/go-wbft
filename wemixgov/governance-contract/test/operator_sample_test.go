package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// TESTS
// 1. single operator - multisig 없이 register, add stake, unstake, claim
// 1-1. 각 amount 값 검증
// 2, withdraw - revert when owner tries to withdraw by .cal() or .transfer() or .send()
// 3. withdraw - withdraw when owner tries to submit tx that .cal() or .transfer() or .send()
// 4. withdraw - withdraw when owner tires to withdraw by right function
// 5. withdraw 후 각 amount 값 검증
// 6. addOwner, removeOwner, change Owner 테스트
// 7. multiple operator - multisig 동작 확인. ( 1~ 5) MultiSig 버전으로 확인

func TestOperatorSampleDeploy(t *testing.T) {
	var (
		operatorContractSingleOwner = NewEOA()
	)
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		operatorContractSingleOwner.Address: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	defer g.backend.Close()
	owners := []*EOA{operatorContractSingleOwner}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, new(big.Int))
	t.Log(operatorSampleAddr)

}
