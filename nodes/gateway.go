package nodes

import (
	"context"
	"fmt"
	"github.com/massbitprotocol/turbo/blockchain"
	"github.com/massbitprotocol/turbo/blockchain/network"
	"github.com/massbitprotocol/turbo/config"
	"github.com/massbitprotocol/turbo/connections"
	"github.com/massbitprotocol/turbo/connections/handler"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"golang.org/x/sync/errgroup"
	"path"
	"strings"
	"time"
)

type gateway struct {
	Base
	context context.Context
	cancel  context.CancelFunc

	bridge          blockchain.Bridge
	blockchainPeers []types.NodeEndpoint
	clock           utils.Clock
	timeStarted     time.Time

	gatewayPeers string
}

var _ connections.TbListener = (*gateway)(nil)

func NewGateway(parent context.Context, tbConfig *config.TurboConfig, bridge blockchain.Bridge,
	blockchainPeers []types.NodeEndpoint, peersInfo []network.PeerInfo) (Node, error) {
	ctx, cancel := context.WithCancel(parent)
	clock := utils.RealClock{}

	g := &gateway{
		Base:            NewBase(tbConfig, "datadir"),
		bridge:          bridge,
		context:         ctx,
		cancel:          cancel,
		blockchainPeers: blockchainPeers,
		clock:           clock,
		timeStarted:     clock.Now(),
		gatewayPeers:    generatePeers(peersInfo),
	}

	return g, nil
}

func generatePeers(peersInfo []network.PeerInfo) string {
	var result string
	if len(peersInfo) == 0 {
		return result
	}
	for _, peer := range peersInfo {
		result += fmt.Sprintf("%s+%s,", peer.Enode.String(), peer.EthWsURI)
	}
	result = result[:len(result)-1]
	return result
}

func (g *gateway) Run() error {
	defer g.cancel()

	var err error

	privateCertDir := path.Join(g.TbConfig.DataDir)
	gatewayType := g.TbConfig.NodeType

	privateCertFile, privateKeyFile := utils.GetCertDir(privateCertDir, strings.ToLower(gatewayType.String()))
	sslCerts := utils.NewSSLCertsFromFiles(privateCertFile, privateKeyFile)

	var group errgroup.Group
	group.Go(g.handleBridgeMessages)

	relayInstructions := make(chan connections.RelayInstruction)
	go g.handleRelayConnections(relayInstructions, sslCerts)
	relayInstructions <- connections.RelayInstruction{
		IP:   "127.0.0.1",
		Type: connections.Connect,
		Port: 443,
	}

	err = group.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (g *gateway) handleRelayConnections(instructions chan connections.RelayInstruction, sslCerts utils.SSLCerts) {
	for {
		instruction := <-instructions
		switch instruction.Type {
		case connections.Connect:
			g.connectRelay(instruction, sslCerts)
		}
	}
}

func (g *gateway) connectRelay(instruction connections.RelayInstruction, sslCerts utils.SSLCerts) {
	relay := handler.NewOutboundRelay(g, &sslCerts, instruction.IP, instruction.Port, "", utils.RealClock{})
	var _ = relay.Start()
	log.Infof("gateway %v (%v) starting, connecting to relay %v:%v", "", g.TbConfig.Environment, instruction.IP, instruction.Port)
}

func (g *gateway) broadcast(msg tbmessage.Message, source connections.Conn, to utils.NodeType) types.BroadcastResults {
	g.ConnectionsLock.RLock()
	defer g.ConnectionsLock.RUnlock()
	results := types.BroadcastResults{}

	for _, conn := range g.Connections {
		// if connection type is not in target - skip
		if conn.Info().ConnectionType&to == 0 {
			continue
		}

		results.RelevantPeers++
		if !conn.IsOpen() || source != nil && conn.ID() == source.ID() {
			results.NotOpenPeers++
			continue
		}

		err := conn.Send(msg)
		if err != nil {
			conn.Log().Errorf("error writing to connection, closing")
			results.ErrorPeers++
			continue
		}

		if conn.Info().IsGateway() {
			results.SentGatewayPeers++
		}

		results.SentPeers++
	}
	return results
}

func (g *gateway) handleBridgeMessages() error {
	for {
		select {
		case txsFromNode := <-g.bridge.ReceiveNodeTransactions():
			blockchainConnection := connections.NewBlockchainConn(txsFromNode.PeerEndpoint)
			for _, blockchainTx := range txsFromNode.Transactions {
				tx := tbmessage.NewTx(blockchainTx.Hash(), blockchainTx.Content(), 1)
				g.processTransaction(tx, blockchainConnection)
			}
		}
	}
}

func (g *gateway) processTransaction(tx *tbmessage.Tx, source connections.Conn) {
	_ = g.broadcast(tx, source, utils.RelayTransaction)
}
