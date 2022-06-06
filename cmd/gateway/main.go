package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/massbitprotocol/turbo/blockchain/eth"
	"github.com/massbitprotocol/turbo/blockchain/network"
	"github.com/massbitprotocol/turbo/config"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"github.com/massbitprotocol/turbo/version"
	"github.com/urfave/cli/v2"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := &cli.App{
		Name:  "gateway",
		Usage: "run a Turbo gateway",
		Flags: []cli.Flag{
			utils.PortFlag,
			utils.EnvFlag,
			utils.LogLevelFlag,
			utils.LogFileLevelFlag,
			utils.LogMaxSizeFlag,
			utils.LogMaxAgeFlag,
			utils.LogMaxBackupsFlag,
			utils.TxTraceEnabledFlag,
			utils.TxTraceMaxFileSizeFlag,
			utils.TxTraceMaxBackupFilesFlag,
			utils.BlockchainNetworkFlag,
			utils.EnodesFlag,
			utils.MultiNode,
			utils.PrivateKeyFlag,
			utils.DisableProfilingFlag,
		},
		Action: runGateway,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runGateway(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !c.Bool(utils.DisableProfilingFlag.Name) {
		go func() {
			log.Infof("pprof http server is running on 0.0.0.0:6060 - %v", "http://localhost:6060/debug/pprof")
			log.Error(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	turboConfig, err := config.NewTurboConfigFromCLI(c)
	if err != nil {
		return err
	}

	ethConfig, err := network.NewPresetEthConfigFromCLI(c)
	if err != nil {
		return err
	}

	var blockchainPeers []types.NodeEndpoint
	for _, peer := range ethConfig.StaticPeers {
		blockchainPeer := peer.Enode
		enodePublicKey := fmt.Sprintf("%x", crypto.FromECDSAPub(blockchainPeer.Pubkey())[1:])
		blockchainPeers = append(blockchainPeers, types.NodeEndpoint{
			IP:        blockchainPeer.IP().String(),
			Port:      blockchainPeer.TCP(),
			PublicKey: enodePublicKey,
		})
	}
	startupBlockchainClient := len(blockchainPeers) > 0

	err = log.Init(turboConfig.Config, version.BuildVersion)
	if err != nil {
		return err
	}

	var blockchainServer *eth.Server
	if startupBlockchainClient {
		log.Infof("starting blockchain client with config for network ID: %v", ethConfig.Network)

		blockchainServer, err = eth.NewServerWithEthLogger(ctx, ethConfig, c.String(utils.DataDirFlag.Name))
		if err != nil {
			return err
		}

		if err = blockchainServer.Start(); err != nil {
			return nil
		}
	} else {
		log.Infof("skipping starting blockchain client as no enodes have been provided")
	}

	<-sigc

	if blockchainServer != nil {
		blockchainServer.Stop()
	}
	return nil
}
