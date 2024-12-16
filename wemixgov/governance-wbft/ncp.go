package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_NCP_LIST = "0x0"
)

func AddNCP(stateDB StateDB, ncp common.Address) error {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Add(stateDB, GovNCPAddress, ncp)
}

func NCPLength(stateDB StateDB) uint64 {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Length(stateDB, GovNCPAddress)
}

func IsNCP(stateDB StateDB, ncp common.Address) bool {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Contains(stateDB, GovNCPAddress, ncp)
}

func NCPList(stateDB StateDB) []common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.Values(stateDB, GovNCPAddress)
}

func NCPAt(stateDB StateDB, index *big.Int) common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_NCP_LIST))
	return ncpSet.At(stateDB, GovNCPAddress, index)
}
