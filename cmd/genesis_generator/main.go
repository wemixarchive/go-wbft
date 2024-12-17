package main

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "genesis generator"
	app.Usage = "generate genesis.json file for your private network consensus engine"
	app.Action = runGenesisGenerator
	app.Run(os.Args)
}

func runGenesisGenerator(c *cli.Context) error {
	network := c.String("network")
	if strings.Contains(network, " ") || strings.Contains(network, "-") || strings.ToLower(network) != network {
		log.Crit("No spaces, hyphens or capital letters allowed in network name")
	}
	makeGenerator(c.String("network")).run()
	return nil
}
