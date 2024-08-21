package wpoa

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/wemixgov"
)

type WemixFakePoA struct {
	wpoa   *WemixPoA
	prvKey *ecdsa.PrivateKey
}

func NewWemixFakeEngine(prvKey *ecdsa.PrivateKey, backend wemixgov.GovBackend) consensus.Engine {
	wkpoa := &WemixFakePoA{
		wpoa: &WemixPoA{
			govCli: NewWemixGov(backend),
		},
		prvKey: prvKey,
	}

	return wkpoa
}

func (wfpoa *WemixFakePoA) Author(header *types.Header) (common.Address, error) {
	return wfpoa.wpoa.Author(header)
}

func (wfpoa *WemixFakePoA) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return wfpoa.wpoa.VerifyHeader(chain, header)
}

func (wfpoa *WemixFakePoA) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	return wfpoa.wpoa.VerifyHeaders(chain, headers)
}

func (wfpoa *WemixFakePoA) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return wfpoa.wpoa.VerifyUncles(chain, block)
}

func (wfpoa *WemixFakePoA) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return wfpoa.wpoa.Prepare(chain, header)
}

func (wfpoa *WemixFakePoA) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) {
	wfpoa.wpoa.Finalize(chain, header, state, txs, uncles, withdrawals)
}

func (wfpoa *WemixFakePoA) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	wfpoa.wpoa.Finalize(chain, header, state, txs, uncles, withdrawals)

	var err error
	header.Coinbase, header.MinerNodeSig, err = wfpoa.signBlock(header.Number, header.Root)
	if err != nil {
		return nil, err
	}
	return types.NewBlock(header, txs, uncles, receipts, trie.NewStackTrie(nil)), nil
}

func (wfpoa *WemixFakePoA) signBlock(height *big.Int, hash common.Hash) (common.Address, []byte, error) {
	sig, err := crypto.Sign(crypto.Keccak256(append(height.Bytes(), hash.Bytes()...)), wfpoa.prvKey)
	if err != nil {
		return common.Address{}, nil, err
	}

	return crypto.PubkeyToAddress(wfpoa.prvKey.PublicKey), sig, nil
}

func (wfpoa *WemixFakePoA) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	return wfpoa.wpoa.Seal(chain, block, results, stop)
}

func (wfpoa *WemixFakePoA) SealHash(header *types.Header) common.Hash {
	return wfpoa.wpoa.SealHash(header)
}

func (wfpoa *WemixFakePoA) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return wfpoa.wpoa.CalcDifficulty(chain, time, parent)
}

func (wfpoa *WemixFakePoA) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return wfpoa.wpoa.APIs(chain)
}

func (wfpoa *WemixFakePoA) Close() error {
	return wfpoa.wpoa.Close()
}
