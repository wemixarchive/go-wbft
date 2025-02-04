package backend

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)

type testEnv struct {
	addrs         []common.Address
	index         map[common.Address]int
	chains        map[common.Address]*core.BlockChain
	engines       map[common.Address]*Backend
	newRoundReady map[common.Address]chan uint64
	results       map[common.Address]chan *types.Block
	parent        *types.Block
	stopCh        chan struct{}

	// properties for test scenario
	down map[common.Address]bool
}

// all engines are waiting for this to be called
func (env *testEnv) waitToSync(t *testing.T, rounds ...uint64) {
	for addr, newRoundReady := range env.newRoundReady {
		result := <-newRoundReady
		if rounds[env.index[addr]] != result {
			t.Errorf("rounds are mismatch: have %d, want %d", result, rounds[env.index[addr]])
		}
	}
}

func (env *testEnv) tryMakeBlock() {
	env.stopCh = make(chan struct{})
	for _, engine := range env.engines {
		chain := env.chains[engine.address]
		block := makeBlockNoNewChainHead(chain, engine, env.parent)
		currState, _ := chain.State()
		block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)

		// all engines try to seal
		engine.Seal(chain, block, env.results[engine.address], env.stopCh)
	}
}

func (env *testEnv) mustSucceed(t *testing.T) *types.Block {
	var result *types.Block
	var proposer common.Address
	for _, addr := range env.addrs {
		go func(addr common.Address) {
			block := <-env.results[addr]
			if block != nil {
				result = block
				proposer = addr
				close(env.stopCh) // stop other `Seal`
			}
		}(addr)
	}
	<-env.stopCh

	if result.ParentHash() != env.parent.Hash() {
		t.Errorf("parent hash mismatch: have %v, want %v", result.ParentHash(), env.parent.Hash())
	}
	if _, err := env.chains[proposer].InsertChain(types.Blocks{result}); err != nil {
		t.Errorf("failed to make block. err %v", err)
		return nil
	}
	env.engines[proposer].NewChainHead() // progress to next sequence
	env.parent = result
	t.Logf("A block is created successfully. proposer=%s\n", proposer.String())
	return result
}

func (env *testEnv) mustGoToNextRound(t *testing.T) {
	// for 1 second, wait for result and check round
}

func (env *testEnv) setEngineDown(index ...int) {
	for _, i := range index {
		env.down[env.addrs[i]] = true
	}
}

func (env *testEnv) setEngineUp(index ...int) {
	for _, i := range index {
		env.down[env.addrs[i]] = false
	}
}

func makeNotifyNewRound(notifyChan chan uint64) func(waitTime time.Duration, round *big.Int) {
	return func(waitTime time.Duration, round *big.Int) {
		notifyChan <- round.Uint64()
	}
}

type simBroadcaster struct {
	env      *testEnv
	myAddr   common.Address
	myChain  *core.BlockChain
	myEngine *Backend
	peers    map[common.Address]consensus.Peer
}

func makeSimBroadcaster(env *testEnv, myAddr common.Address) *simBroadcaster {
	peers := make(map[common.Address]consensus.Peer)
	for addr, engine := range env.engines {
		if addr == myAddr {
			continue
		}
		peers[addr] = &simPeer{
			myAddr:     myAddr,
			peerAddr:   addr,
			peerEngine: engine,
		}
	}
	return &simBroadcaster{
		env:      env,
		myAddr:   myAddr,
		myChain:  env.chains[myAddr],
		myEngine: env.engines[myAddr],
		peers:    peers,
	}
}

func (sb *simBroadcaster) Enqueue(id string, block *types.Block) {
	sb.myChain.InsertChain(types.Blocks{block})
	sb.myEngine.NewChainHead() // progress to next sequence
}

func (sb *simBroadcaster) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	if sb.env.down[sb.myAddr] {
		// my engine is down!
		return m
	}
	for addr, p := range sb.peers {
		if targets[addr] && !sb.env.down[addr] {
			m[addr] = p
		}
	}
	return m
}

type simPeer struct {
	myAddr     common.Address
	peerAddr   common.Address
	peerEngine *Backend
}

func (sp *simPeer) SendQBFTConsensus(msgcode uint64, payload []byte) error {
	if err := sp.peerEngine.istanbulEventMux.Post(qbft.MessageEvent{
		Code:    msgcode,
		Payload: payload,
	}); err != nil {
		return err
	}
	return nil
}

// make n engines with given genesis and cfg
func makeMultiEngineTestEnv(n int) (env *testEnv) {
	env = &testEnv{}

	// validators are ordered so the proposer will be selected in index order
	genesis, nodeKeys := testutils.GenesisAndKeys(n)
	env.parent = genesis.ToBlock()
	env.down = make(map[common.Address]bool)
	env.results = make(map[common.Address]chan *types.Block)

	config := new(qbft.Config)
	setConfigFromChainConfig(config, genesis.Config)
	config.BlockPeriod = 1
	config.RequestTimeout = 2000
	config.MaxRequestTimeoutSeconds = 2
	config.AllowedFutureBlockTime = 100000000 // to skip future block check; this makes block creation time to be very short

	env.addrs = make([]common.Address, n)
	env.index = make(map[common.Address]int)
	env.chains = make(map[common.Address]*core.BlockChain)
	env.engines = make(map[common.Address]*Backend)
	for i, nodeKey := range nodeKeys {
		addr := crypto.PubkeyToAddress(nodeKey.PublicKey)
		memDB := rawdb.NewMemoryDatabase()
		env.addrs[i] = addr
		env.index[addr] = i
		env.engines[addr] = New(config, nodeKey, memDB)
		env.results[addr] = make(chan *types.Block)

		genesis.MustCommit(memDB, triedb.NewDatabase(memDB, triedb.HashDefaults))

		var err error
		env.chains[addr], err = core.NewBlockChain(memDB, nil, genesis, nil, env.engines[addr], vm.Config{}, nil, nil)
		if err != nil {
			panic(err)
		}
	}
	env.newRoundReady = make(map[common.Address]chan uint64)
	for addr, engine := range env.engines {
		engine.broadcaster = makeSimBroadcaster(env, addr)
		env.newRoundReady[addr] = make(chan uint64)

		// engine tries to `NotifyNewRound` at `Start`, so it will be blocked until we call `waitToSync`
		go engine.Start(env.chains[addr], env.chains[addr].CurrentFullBlock, rawdb.HasBadBlock, makeNotifyNewRound(env.newRoundReady[addr]))
	}
	return
}

func makeBlockNoNewChainHead(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	header := makeHeader(chain.Config(), engine.config, parent)
	engine.Prepare(chain, header)
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}

func TestWBFTSimpleCase(t *testing.T) {
	env := makeMultiEngineTestEnv(4)

	for i := 0; i < 3; i++ {
		// wait for all engine to ready for new round
		env.waitToSync(t, 0, 0, 0, 0)
		env.tryMakeBlock()
		env.mustSucceed(t)
	}
}

func TestWBFTOneEngineDown(t *testing.T) {
	env := makeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.waitToSync(t, 0, 0, 0, 0)
	env.tryMakeBlock()
	env.mustSucceed(t)

	// make second block with 3 normal engine and 1 disconnected engine
	env.setEngineDown(3)
	env.waitToSync(t, 0, 0, 0, 0)
	env.tryMakeBlock()
	env.mustSucceed(t)
	env.waitToSync(t, 0, 0, 0, 1) // engine 3 does not receive a block and changes round
}

func TestWBFTTwoEngineDown(t *testing.T) {
	env := makeMultiEngineTestEnv(4)
	env.waitToSync(t, 0, 0, 0, 0)

	// make first block with 4 normal engine
	env.tryMakeBlock()
	env.mustSucceed(t) // proposer is engine 0
	env.waitToSync(t, 0, 0, 0, 0)

	// make second block with 2 normal engine and 2 disconnected engine
	env.setEngineDown(2, 3)
	env.tryMakeBlock()
	env.waitToSync(t, 1, 1, 1, 1)
	t.Log("round changed to 1")

	env.waitToSync(t, 2, 2, 2, 2)
	t.Log("round changed to 2")

	// engine 2 is up
	env.setEngineUp(2)

	env.waitToSync(t, 3, 3, 3, 3)
	t.Log("round changed to 3")

	env.mustSucceed(t) // proposer is engine 0 again after circulation

	t.Log("round changed to 0")
	env.waitToSync(t, 0, 0, 0, 4)
}
