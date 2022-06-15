package nodes

import (
	"context"
	"fmt"
	"github.com/massbitprotocol/turbo/blockchain"
	"github.com/massbitprotocol/turbo/blockchain/network"
	"github.com/massbitprotocol/turbo/config"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"time"
)

type gateway struct {
	context context.Context
	cancel  context.CancelFunc

	bridge          blockchain.Bridge
	blockchainPeers []types.NodeEndpoint
	clock           utils.Clock
	timeStarted     time.Time

	gatewayPeers string
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

func NewGateway(parent context.Context, config *config.TurboConfig, bridge blockchain.Bridge,
	blockchainPeers []types.NodeEndpoint, peersInfo []network.PeerInfo) (Node, error) {
	ctx, cancel := context.WithCancel(parent)
	clock := utils.RealClock{}

	g := &gateway{
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

func (g *gateway) Run() error {
	return nil
}

func (g *gateway) handleBridgeMessages() error {
	return nil
}
