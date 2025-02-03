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
	addrs           []common.Address
	index           map[common.Address]int
	chains          map[common.Address]*core.BlockChain
	engines         map[common.Address]*Backend
	newRoundReady   map[common.Address]chan uint64
	parent          *types.Block
	result          chan *types.Block
	currentProposer common.Address

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
	for _, engine := range env.engines {
		env.currentProposer = engine.ProposerFromValSet()
		break
	}
	block := makeBlockNoNewChainHead(env.chains[env.currentProposer], env.engines[env.currentProposer], env.parent)
	currState, _ := env.chains[env.currentProposer].State()
	block, _ = env.engines[env.currentProposer].FinalizeAndAssemble(env.chains[env.currentProposer], block.Header(), currState, nil, nil, nil, nil)
	stopCh := make(chan struct{})
	env.engines[env.currentProposer].Seal(env.chains[env.currentProposer], block, env.result, stopCh)
}

func (env *testEnv) mustSucceed(t *testing.T) *types.Block {
	block := <-env.result

	if block.ParentHash() != env.parent.Hash() {
		t.Errorf("parent hash mismatch: have %v, want %v", block.ParentHash(), env.parent.Hash())
	}
	if _, err := env.chains[env.currentProposer].InsertChain(types.Blocks{block}); err != nil {
		t.Errorf("failed to make block. err %v", err)
		return nil
	}
	env.engines[env.currentProposer].NewChainHead() // progress to next sequence
	env.parent = block

	return block
}

func (env *testEnv) mustGoToNextRound(t *testing.T) {
	// for 1 second, wait for result and check round
}

func (env *testEnv) setScenarioEngineDown(index ...int) {
	for _, i := range index {
		env.down[env.addrs[i]] = true
	}
}

func (env *testEnv) setScenarioEngineUp(index ...int) {
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
	genesis, nodeKeys := testutils.GenesisAndKeys(n)
	env.parent = genesis.ToBlock()
	env.down = make(map[common.Address]bool)
	env.result = make(chan *types.Block)

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

	// make second block with 3 normal engine and 1 down engine
	env.setScenarioEngineDown(3)
	env.waitToSync(t, 0, 0, 0, 0)
	env.tryMakeBlock()
	env.mustSucceed(t)
	env.waitToSync(t, 0, 0, 0, 0)
}

func TestWBFTTwoEngineDown(t *testing.T) {
	env := makeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.waitToSync(t, 0, 0, 0, 0)
	env.tryMakeBlock()
	env.mustSucceed(t)
	env.waitToSync(t, 0, 0, 0, 0)

	// make second block with 3 normal engine and 1 down engine
	env.setScenarioEngineDown(2, 3)
	env.tryMakeBlock()
	env.waitToSync(t, 1, 1, 1, 1)
	t.Log("round changed to 1")

	env.waitToSync(t, 2, 2, 2, 2)
	t.Log("round changed to 2")

	// engine 2 is up
	env.setScenarioEngineUp(2)

	env.waitToSync(t, 3, 3, 3, 3)
	t.Log("round changed to 3")

	env.mustSucceed(t)

	t.Log("round changed to 0")
	env.waitToSync(t, 0, 0, 0, 0)
}
