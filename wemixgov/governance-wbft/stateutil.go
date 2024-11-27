package governancewbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

type StateDB interface {
	GetState(addr common.Address, hash common.Hash) common.Hash
}

func CalculateMappingSlot(baseSlot common.Hash, key interface{ Bytes() []byte }) common.Hash {
	// keccak256(encode(key) . encode(slot))
	hash := sha3.NewLegacyKeccak256()

	// 키 (주소)와 슬롯 번호를 각각 32바이트로 변환 후 연결
	keyBytes := append(common.LeftPadBytes(key.Bytes(), 32), baseSlot.Bytes()...)
	hash.Write(keyBytes)
	return common.BytesToHash(hash.Sum(nil))
}

func CalculateDynamicSlot(baseSlot interface{ Bytes() []byte }, index *big.Int) common.Hash {
	// keccak256(baseSlot)으로 배열의 시작 위치를 계산
	hash := sha3.NewLegacyKeccak256()
	hash.Write(common.LeftPadBytes(baseSlot.Bytes(), 32))
	arrayStartSlot := new(big.Int).SetBytes(hash.Sum(nil))

	// 배열 요소 슬롯: arrayStartSlot + index
	elementSlot := new(big.Int).Add(arrayStartSlot, index)

	return common.BigToHash(elementSlot)
}

func IncrementHash(baseSlot common.Hash, increment *big.Int) common.Hash {
	return common.BigToHash(new(big.Int).Add(baseSlot.Big(), increment))
}

func HashToAddress(hash common.Hash) common.Address {
	return common.BytesToAddress(hash.Bytes())
}

type EnumerableSet[T interface{ Bytes() []byte }] struct {
	indexSlot   common.Hash
	valueSlot   common.Hash
	convertFunc func(common.Hash) T
}

func NewAddressSet(baseSlot common.Hash) *EnumerableSet[common.Address] {
	es := NewEnumerableSet[common.Address](baseSlot)
	es.convertFunc = HashToAddress
	return es
}

func NewEnumerableSet[T interface{ Bytes() []byte }](baseSlot common.Hash) *EnumerableSet[T] {
	return &EnumerableSet[T]{
		valueSlot: baseSlot,
		indexSlot: IncrementHash(baseSlot, big.NewInt(1)),
	}
}

func (es *EnumerableSet[T]) Length(stateDB StateDB, address common.Address) uint64 {
	return stateDB.GetState(address, es.valueSlot).Big().Uint64()
}

func (es *EnumerableSet[T]) Contains(stateDB StateDB, address common.Address, value T) bool {
	index := stateDB.GetState(address, CalculateMappingSlot(es.indexSlot, value)).Big()

	return index.Sign() > 0
}

func (es *EnumerableSet[T]) Values(stateDB StateDB, address common.Address) []T {
	len := es.Length(stateDB, address)
	values := make([]T, len)
	for i := uint64(0); i < len; i++ {
		values[i] = es.convertFunc(stateDB.GetState(address, CalculateDynamicSlot(es.valueSlot, new(big.Int).SetUint64(i))))
	}
	return values
}

func (es *EnumerableSet[T]) At(stateDB StateDB, address common.Address, index *big.Int) T {
	return es.convertFunc(stateDB.GetState(address, CalculateDynamicSlot(es.valueSlot, index)))
}
