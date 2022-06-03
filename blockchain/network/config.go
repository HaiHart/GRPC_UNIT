package network

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"math/big"
	"time"
)

// PeerInfo contains the enode and websocket endpoint for an Ethereum peer
type PeerInfo struct {
	Enode    *enode.Node
	EthWsURI string
}

// EthConfig represents Ethereum network configuration settings (e.g. indicate Mainnet, Rinkeby, BSC, etc.)
type EthConfig struct {
	StaticPeers []PeerInfo
	PrivateKey  *ecdsa.PrivateKey

	Network                 uint64
	TotalDifficulty         *big.Int
	Head                    common.Hash
	Genesis                 common.Hash
	BlockConfirmationsCount int
	SendBlockConfirmation   bool

	IgnoreBlockTimeout time.Duration
}

// StaticEnodes makes a list of enodes only from StaticPeers
func (ec *EthConfig) StaticEnodes() []*enode.Node {
	enodes := make([]*enode.Node, 0, len(ec.StaticPeers))
	for _, peer := range ec.StaticPeers {
		enodes = append(enodes, peer.Enode)
	}
	return enodes
}
