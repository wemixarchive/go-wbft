package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		os.Exit(0)
	}()

	app := cli.NewApp()
	app.Name = "genesis generator"
	app.Usage = "generate genesis.json file for your private network consensus engine"
	app.Action = runGenesisGenerator
	app.Run(os.Args)
}

func runGenesisGenerator(c *cli.Context) error {
	makeGenerator(c.String("network")).run()
	return nil
}
