package network

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/massbitprotocol/turbo/utils"
	"github.com/urfave/cli/v2"
	"math/big"
	"time"
)

const privateKeyLen = 64

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

// NewPresetEthConfigFromCLI builds a new EthConfig from the command line context. Selects a specific network configuration based on the provided startup flag.
func NewPresetEthConfigFromCLI(ctx *cli.Context) (*EthConfig, error) {
	preset, err := NewEthereumPreset(ctx.String(utils.BlockchainNetworkFlag.Name))
	if err != nil {
		return nil, err
	}

	var peers []PeerInfo
	if ctx.IsSet(utils.MultiNode.Name) {

	} else if ctx.IsSet(utils.EnodesFlag.Name) {
		enodeStr := ctx.String(utils.EnodesFlag.Name)
		peers = make([]PeerInfo, 0, 1)
		var peer PeerInfo

		node, err := enode.Parse(enode.ValidSchemes, enodeStr)
		if err != nil {
			return nil, err
		}
		peer.Enode = node

		peers = append(peers, peer)
	}

	preset.StaticPeers = peers

	if ctx.IsSet(utils.PrivateKeyFlag.Name) {
		privateKeyHexString := ctx.String(utils.PrivateKeyFlag.Name)
		if len(privateKeyHexString) != privateKeyLen {
			return nil, fmt.Errorf("incorrect private key length: expected length %v, actual length %v", privateKeyLen, len(privateKeyHexString))
		}
		privateKey, err := crypto.HexToECDSA(privateKeyHexString)
		if err != nil {
			return nil, err
		}
		preset.PrivateKey = privateKey
	}

	return &preset, nil
}

// StaticEnodes makes a list of enodes only from StaticPeers
func (ec *EthConfig) StaticEnodes() []*enode.Node {
	enodes := make([]*enode.Node, 0, len(ec.StaticPeers))
	for _, peer := range ec.StaticPeers {
		enodes = append(enodes, peer.Enode)
	}
	return enodes
}
