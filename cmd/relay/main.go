package main

import (
	"github.com/massbitprotocol/turbo/nodes"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:   "gateway",
		Usage:  "run a Turbo relay",
		Flags:  []cli.Flag{},
		Action: runRelay,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runRelay(c *cli.Context) error {
	relay, err := nodes.NewRelay()
	if err != nil {
		return err
	}
	go func() {
		err = relay.Run()
		if err != nil {

		}
	}()
	return nil
}
