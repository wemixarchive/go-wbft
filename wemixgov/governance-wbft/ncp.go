package governancewbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_NCP_LIST = "0x0"
)

type GovNCP struct {
	Address common.Address
	ncpSet  *EnumerableSet[common.Address]
}

func NewGovNCP(address common.Address) *GovNCP {
	if address == (common.Address{}) {
		return nil
	}
	return &GovNCP{
		Address: address,
		ncpSet:  NewAddressSet(common.HexToHash(SLOT_NCP_LIST)),
	}
}

func (gn *GovNCP) NCPLength(stateDB StateDB) uint64 {
	return gn.ncpSet.Length(stateDB, gn.Address)
}

func (gn *GovNCP) IsNCP(stateDB StateDB, ncp common.Address) bool {
	return gn.ncpSet.Contains(stateDB, gn.Address, ncp)
}

func (gn *GovNCP) NCPList(stateDB StateDB) []common.Address {
	return gn.ncpSet.Values(stateDB, gn.Address)
}

func (gn *GovNCP) NCPAt(stateDB StateDB, index *big.Int) common.Address {
	return gn.ncpSet.At(stateDB, gn.Address, index)
}
