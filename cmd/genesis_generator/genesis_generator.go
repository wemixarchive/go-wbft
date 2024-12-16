package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/peterh/liner"
	"golang.org/x/term"
)

type config struct {
	Genesis *core.Genesis `json:"genesis,omitempty"` // Genesis block to cache for node deploys
}

type genesisGenerator struct {
	conf config
}

func newUint64(val uint64) *uint64 { return &val }

func makeGenerator(network string) *genesisGenerator {
	return &genesisGenerator{
		conf: config{},
	}
}

func (g *genesisGenerator) run() {
	fmt.Println("+-----------------------------------------------------------+")
	fmt.Println("| Genesis Generator is a tool that generates an genesis file  |")
	fmt.Println("| according to the desired consensus engines in wemix chain   |")
	fmt.Println("| from your cli inputs.                                       |")
	fmt.Println("| Don't wander through vast docs, just simply generate it!    |")
	fmt.Println("+-----------------------------------------------------------+")
	fmt.Println()

	g.makeGenesis()
}

func (g *genesisGenerator) makeGenesis() {
	// Construct a default genesis block
	genesis := &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   4700000,
		Difficulty: big.NewInt(524288),
		Alloc:      make(types.GenesisAlloc),
		Config: &params.ChainConfig{
			HomesteadBlock:      big.NewInt(0),
			EIP150Block:         big.NewInt(0),
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
			MuirGlacierBlock:    big.NewInt(0),
			BerlinBlock:         big.NewInt(0),
			LondonBlock:         big.NewInt(0),
			ArrowGlacierBlock:   big.NewInt(0),
			GrayGlacierBlock:    big.NewInt(0),
		},
	}
	// Figure out which consensus engine to choose
	fmt.Println()
	fmt.Println("Which consensus engine to use? (default = Wemix)")
	fmt.Println(" 1. Ethash - PoW")
	fmt.Println(" 2. Beacon Ethash - beacon engine switched from ethash")
	fmt.Println(" 3. Clique - PoA")
	fmt.Println(" 4. Beacon Clique - beacon engine switched from clique")
	fmt.Println(" 5. Wbft - wemix DPoS")
	fmt.Println(" 6. Beacon Wbft - beacon engine switched from wbft")
	fmt.Println(" 7. Wemix - wemix engine swtiched from wpoa to wbft")

	choice := g.read()
	switch {
	case choice == "1":
		g.ethashConfig(genesis)

	case choice == "2":
		g.beaconChainConfig(genesis)
		g.ethashConfig(genesis)

	case choice == "3":
		g.cliqueConfig(genesis)

	case choice == "4":
		g.beaconChainConfig(genesis)
		g.cliqueConfig(genesis)

	default:
		log.Crit("Invalid consensus engine choice", "choice", choice)
	}
	// Consensus all set, just ask for initial funds and go
	fmt.Println()
	fmt.Println("Which accounts should be pre-funded? (advisable at least one)")
	for {
		// Read the address of the account to fund
		if address := g.readAddress(); address != nil {
			genesis.Alloc[*address] = types.Account{
				Balance: new(big.Int).Lsh(big.NewInt(1), 256-7), // 2^256 / 128 (allow many pre-funds without balance overflows)
			}
			continue
		}
		break
	}

	// Query the user for some custom extras
	fmt.Println()
	fmt.Println("Specify your chain/network ID if you want an explicit one (default = random)")
	genesis.Config.ChainID = new(big.Int).SetUint64(uint64(g.readDefaultInt(rand.Intn(65536))))

	// All done, store the genesis and flush to disk
	log.Info("Configured new genesis block")
	g.conf.Genesis = genesis

	fmt.Println()
	fmt.Println(" Do you want to export generated genesis file?")
	fmt.Println(" 1. yes")
	fmt.Println(" 2. nah, just print it")

	choice = g.read()
	switch {
	case choice == "1":
		fmt.Println()
		fmt.Printf("Which folder to save the genesis spec into? (default = current)\n")
		fmt.Printf("It will create genesis.json\n")

		folder := g.readDefaultString(".")
		g.flush(folder)

	case choice == "2":
		g.flush("")
	default:
		g.flush("")
	}
}

func (g *genesisGenerator) beaconChainConfig(genesis *core.Genesis) {
	fmt.Println()
	fmt.Println("Do you want to start beacon chain immediately? (default yes)")
	if g.readDefaultYesNo(true) {
		genesis.Config.TerminalTotalDifficulty = common.Big0
		genesis.Config.TerminalTotalDifficultyPassed = true
		genesis.Config.ShanghaiTime = newUint64(0)
		genesis.Config.CancunTime = newUint64(0)

	} else {
		genesis.Config.TerminalTotalDifficultyPassed = false
		fmt.Println()
		fmt.Println("Enter TerminalTotalDifficulty value you want to set (default 58_750_000_000_000_000_000_000)")
		ttd := g.readDefaultBigInt(params.MainnetTerminalTotalDifficulty)
		genesis.Config.TerminalTotalDifficulty = ttd
		fmt.Println()
		fmt.Println("Enter timestamp you want to enable Shanghai Fork (default 1677557088)")
		shanghaiTime := g.readDefaultInt(1677557088)
		genesis.Config.ShanghaiTime = newUint64(uint64(shanghaiTime))
		fmt.Println()
		fmt.Println("Enter timestamp you want to enable Cancun Fork (default 1706655072)")
		cancunTime := g.readDefaultInt(1706655072)
		genesis.Config.CancunTime = newUint64(uint64(cancunTime))
	}
}

func (g *genesisGenerator) ethashConfig(genesis *core.Genesis) {
	genesis.Config.Ethash = new(params.EthashConfig)
	genesis.ExtraData = make([]byte, 32)
}

func (g *genesisGenerator) cliqueConfig(genesis *core.Genesis) {
	genesis.Difficulty = big.NewInt(1)
	genesis.Config.Clique = &params.CliqueConfig{
		Period: 15,
		Epoch:  30000,
	}
	fmt.Println()
	fmt.Println("How many seconds should blocks take? (default = 15)")
	genesis.Config.Clique.Period = uint64(g.readDefaultInt(15))

	// We also need the initial list of signers
	fmt.Println()
	fmt.Println("Which accounts are allowed to seal? (mandatory at least one)")

	var signers []common.Address
	for {
		if address := g.readAddress(); address != nil {
			signers = append(signers, *address)
			continue
		}
		if len(signers) > 0 {
			break
		}
	}
	// Sort the signers and embed into the extra-data section
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if bytes.Compare(signers[i][:], signers[j][:]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	genesis.ExtraData = make([]byte, 32+len(signers)*common.AddressLength+65)
	for i, signer := range signers {
		copy(genesis.ExtraData[32+i*common.AddressLength:], signer[:])
	}
}

// prompts the user for input with the given prompt string.  Returns when a value is entered.
// Causes the genesisGenerator to exit if ctrl-d is pressed
func promptInput(p string) string {
	for {
		text, err := prompt.Stdin.PromptInput(p)
		if err != nil {
			if err != liner.ErrPromptAborted {
				log.Crit("Failed to read user input", "err", err)
			}
		} else {
			return text
		}
	}
}

// read reads a single line from stdin, trimming if from spaces.
func (g *genesisGenerator) read() string {
	text := promptInput("> ")
	return strings.TrimSpace(text)
}

// readString reads a single line from stdin, trimming if from spaces, enforcing
// non-emptyness.
func (g *genesisGenerator) readString() string {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text != "" {
			return text
		}
	}
}

// readDefaultString reads a single line from stdin, trimming if from spaces. If
// an empty line is entered, the default value is returned.
func (g *genesisGenerator) readDefaultString(def string) string {
	text := promptInput("> ")
	if text = strings.TrimSpace(text); text != "" {
		return text
	}
	return def
}

// readDefaultYesNo reads a single line from stdin, trimming if from spaces and
// interpreting it as a 'yes' or a 'no'. If an empty line is entered, the default
// value is returned.
func (g *genesisGenerator) readDefaultYesNo(def bool) bool {
	for {
		text := promptInput("> ")
		if text = strings.ToLower(strings.TrimSpace(text)); text == "" {
			return def
		}
		if text == "y" || text == "yes" {
			return true
		}
		if text == "n" || text == "no" {
			return false
		}
		log.Error("Invalid input, expected 'y', 'yes', 'n', 'no' or empty")
	}
}

// readURL reads a single line from stdin, trimming if from spaces and trying to
// interpret it as a URL (http, https or file).
func (g *genesisGenerator) readURL() *url.URL {
	for {
		text := promptInput("> ")
		uri, err := url.Parse(strings.TrimSpace(text))
		if err != nil {
			log.Error("Invalid input, expected URL", "err", err)
			continue
		}
		return uri
	}
}

// readInt reads a single line from stdin, trimming if from spaces, enforcing it
// to parse into an integer.
func (g *genesisGenerator) readInt() int {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			continue
		}
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			log.Error("Invalid input, expected integer", "err", err)
			continue
		}
		return val
	}
}

// readDefaultInt reads a single line from stdin, trimming if from spaces, enforcing
// it to parse into an integer. If an empty line is entered, the default value is
// returned.
func (g *genesisGenerator) readDefaultInt(def int) int {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		val, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			log.Error("Invalid input, expected integer", "err", err)
			continue
		}
		return val
	}
}

// readDefaultBigInt reads a single line from stdin, trimming if from spaces,
// enforcing it to parse into a big integer. If an empty line is entered, the
// default value is returned.
func (g *genesisGenerator) readDefaultBigInt(def *big.Int) *big.Int {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		val, ok := new(big.Int).SetString(text, 0)
		if !ok {
			log.Error("Invalid input, expected big integer")
			continue
		}
		return val
	}
}

// readDefaultFloat reads a single line from stdin, trimming if from spaces, enforcing
// it to parse into a float. If an empty line is entered, the default value is returned.
func (g *genesisGenerator) readDefaultFloat(def float64) float64 {
	for {
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			log.Error("Invalid input, expected float", "err", err)
			continue
		}
		return val
	}
}

// readPassword reads a single line from stdin, trimming it from the trailing new
// line and returns it. The input will not be echoed.
func (g *genesisGenerator) readPassword() string {
	fmt.Printf("> ")
	text, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Crit("Failed to read password", "err", err)
	}
	fmt.Println()
	return string(text)
}

// readAddress reads a single line from stdin, trimming if from spaces and converts
// it to an Ethereum address.
func (g *genesisGenerator) readAddress() *common.Address {
	for {
		text := promptInput("> 0x")
		if text = strings.TrimSpace(text); text == "" {
			return nil
		}
		// Make sure it looks ok and return it if so
		if len(text) != 40 {
			log.Error("Invalid address length, please retry")
			continue
		}
		bigaddr, _ := new(big.Int).SetString(text, 16)
		address := common.BigToAddress(bigaddr)
		return &address
	}
}

// readDefaultAddress reads a single line from stdin, trimming if from spaces and
// converts it to an Ethereum address. If an empty line is entered, the default
// value is returned.
func (g *genesisGenerator) readDefaultAddress(def common.Address) common.Address {
	for {
		// Read the address from the user
		text := promptInput("> 0x")
		if text = strings.TrimSpace(text); text == "" {
			return def
		}
		// Make sure it looks ok and return it if so
		if len(text) != 40 {
			log.Error("Invalid address length, please retry")
			continue
		}
		bigaddr, _ := new(big.Int).SetString(text, 16)
		return common.BigToAddress(bigaddr)
	}
}

// readJSON reads a raw JSON message and returns it.
func (g *genesisGenerator) readJSON() string {
	var blob json.RawMessage

	for {
		text := promptInput("> ")
		reader := strings.NewReader(text)
		if err := json.NewDecoder(reader).Decode(&blob); err != nil {
			log.Error("Invalid JSON, please try again", "err", err)
			continue
		}
		return string(blob)
	}
}

// readIPAddress reads a single line from stdin, trimming if from spaces and
// returning it if it's convertible to an IP address. The reason for keeping
// the user input format instead of returning a Go net.IP is to match with
// weird formats used by ethstats, which compares IPs textually, not by value.
func (g *genesisGenerator) readIPAddress() string {
	for {
		// Read the IP address from the user
		fmt.Printf("> ")
		text := promptInput("> ")
		if text = strings.TrimSpace(text); text == "" {
			return ""
		}
		// Make sure it looks ok and return it if so
		if ip := net.ParseIP(text); ip == nil {
			log.Error("Invalid IP address, please retry")
			continue
		}
		return text
	}
}

// flush dumps the contents of config to disk or print.
func (g *genesisGenerator) flush(folder string) {
	out, _ := json.MarshalIndent(g.conf.Genesis, "", "  ")

	if folder != "" {
		if err := os.MkdirAll(folder, 0755); err != nil {
			log.Error("Failed to create spec folder", "folder", folder, "err", err)
			return
		}
		gethJson := filepath.Join(folder, fmt.Sprintf("genesis.json"))
		if err := os.WriteFile(gethJson, out, 0644); err != nil {
			log.Error("Failed to save genesis file", "err", err)
			return
		}
		log.Info("Saved native genesis chain spec", "path", gethJson)
	} else {
		fmt.Println(string(out))
	}

}
