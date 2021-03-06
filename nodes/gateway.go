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
	"github.com/massbitprotocol/turbo/services"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"golang.org/x/sync/errgroup"
	"path"
	"strings"
	"time"
)

const (
	timeToAvoidReEntry = 6 * time.Hour
)

type gateway struct {
	Base
	context         context.Context
	cancel          context.CancelFunc
	bridge          blockchain.Bridge
	asyncMsgChannel chan services.MsgInfo
	blockProcessor  services.BlockProcessor
	blockchainPeers []types.NodeEndpoint
	clock           utils.Clock
	timeStarted     time.Time
	gatewayPeers    string
}

var _ connections.TbListener = (*gateway)(nil)

func NewGateway(parent context.Context, config *config.TurboConfig, bridge blockchain.Bridge, blockchainPeers []types.NodeEndpoint, peersInfo []network.PeerInfo) (Node, error) {
	ctx, cancel := context.WithCancel(parent)
	clock := utils.RealClock{}
	g := &gateway{
		Base:            NewBase(config, "datadir"),
		bridge:          bridge,
		context:         ctx,
		cancel:          cancel,
		blockchainPeers: blockchainPeers,
		clock:           clock,
		timeStarted:     clock.Now(),
		gatewayPeers:    generatePeers(peersInfo),
	}
	g.asyncMsgChannel = services.NewAsyncMsgChannel(g)
	assigner := services.NewEmptyTxIDAssigner()
	txStore := services.NewTbTxStore(30*time.Minute, 3*24*time.Hour, 10*time.Minute, assigner, services.NewHashHistory("seenTxs", 30*time.Minute), nil, timeToAvoidReEntry)
	g.TxStore = &txStore
	g.blockProcessor = services.NewRLPBlockProcessor(g.TxStore)
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

	privateCertDir := path.Join(g.Config.DataDir)
	gatewayType := g.Config.NodeType

	privateCertFile, privateKeyFile := utils.GetCertDir(privateCertDir, strings.ToLower(gatewayType.String()))
	sslCerts := utils.NewSSLCertsFromFiles(privateCertFile, privateKeyFile)

	var group errgroup.Group
	group.Go(g.handleBridgeMessages)
	group.Go(g.TxStore.Start)

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
	relay := handler.NewOutboundRelay(g, &sslCerts, instruction.IP, instruction.Port, "", utils.Relay, utils.RealClock{})
	var _ = relay.Start()
	log.Infof("gateway %v (%v) starting, connecting to relay %v:%v", "", g.Config.Environment, instruction.IP, instruction.Port)
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
		case blockFromNode := <-g.bridge.ReceiveBlockFromNode():
			blockchainConnection := connections.NewBlockchainConn(blockFromNode.PeerEndpoint)
			g.processBlockFromBlockchain(blockFromNode.Block, blockchainConnection)
		}
	}
}

func (g *gateway) HandleMsg(msg tbmessage.Message, source connections.Conn, background connections.MsgHandlingOptions) error {
	var err error
	if background {
		g.asyncMsgChannel <- services.MsgInfo{Msg: msg, Source: source}
		return nil
	}

	switch msg.(type) {
	case *tbmessage.Tx:
		tx := msg.(*tbmessage.Tx)
		g.processTransaction(tx, source)
	default:
		err = g.Base.HandleMsg(msg, source)
	}
	return err
}

func (g *gateway) processTransaction(tx *tbmessage.Tx, source connections.Conn) {
	startTime := time.Now()
	networkDuration := startTime.Sub(tx.Timestamp()).Microseconds()
	sentToBlockchainNode := false
	sentToBDN := false
	// we add the transaction to TxStore with current time, so we can measure time difference to node announcement/confirmation
	txResult := g.TxStore.Add(tx.Hash(), tx.Content(), tx.TxID(), tx.NetworkNum(), g.clock.Now())
	if txResult.NewContent || txResult.Reprocess {
		if txResult.NewContent {

		}

		if !source.Info().IsRelay() {
			if !txResult.Reprocess {

			}

			tx.SetSender(txResult.Sender)
			tx.SetTimestamp(g.clock.Now())
			_ = g.broadcast(tx, source, utils.RelayTransaction)
			sentToBDN = true
		}

		if source.Info().ConnectionType != utils.Blockchain {
			_ = g.bridge.SendTransactionsFromBDN([]*types.TbTransaction{txResult.Transaction})
			sentToBlockchainNode = true
		}
	}

	log.Tracef("msgTx: from %v, hash %v, new Tx %v, new content %v, new TxID %v, sentToBDN: %v, sentToBlockchainNode: %v, handling duration %v, sender %v, networkDuration %v", source, tx.Hash(), txResult.NewTx, txResult.NewContent, txResult.NewTxID, sentToBDN, sentToBlockchainNode, time.Since(startTime), tx.Sender(), networkDuration)
}

func (g *gateway) processBlockFromBlockchain(tbBlock *types.TbBlock, source connections.Blockchain) {
	broadcastMessage, _, _ := g.blockProcessor.BxBlockToBroadcast(tbBlock, 0)
	_ = g.broadcast(broadcastMessage, source, utils.RelayBlock)
}
