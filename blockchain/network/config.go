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
	"strings"
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
		if ctx.IsSet(utils.EnodesFlag.Name) {
			return nil, fmt.Errorf("parameters --multi-node and --enodes should not be used simultaneously; if you would like to use multiple p2p or websockets connections, use --multi-node")
		}
		peers, err = ParseMultiNode(ctx.String(utils.MultiNode.Name))
		if err != nil {
			return nil, err
		}
	} else if ctx.IsSet(utils.EnodesFlag.Name) {
		enodeStr := ctx.String(utils.EnodesFlag.Name)
		peers = make([]PeerInfo, 0, 1)
		var peer PeerInfo

		node, err := enode.Parse(enode.ValidSchemes, enodeStr)
		if err != nil {
			return nil, err
		}
		peer.Enode = node

		if ctx.IsSet(utils.EthWsUriFlag.Name) {
			ethWsURI := ctx.String(utils.EthWsUriFlag.Name)
			err = validateWsURI(ethWsURI)
			if err != nil {
				return nil, err
			}
			peer.EthWsURI = ethWsURI
		}

		peers = append(peers, peer)
	}

	preset.StaticPeers = peers

	if ctx.IsSet(utils.PrivateKeyFlag.Name) {
		privateKeyHex := ctx.String(utils.PrivateKeyFlag.Name)
		if len(privateKeyHex) != privateKeyLen {
			return nil, fmt.Errorf("incorrect private key length: expected length %v, actual length %v", privateKeyLen, len(privateKeyHex))
		}
		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, err
		}
		preset.PrivateKey = privateKey
	}

	return &preset, nil
}

// ParseMultiNode parses a string into list of PeerInfo, according to expected format of multi-eth-ws-uri parameter
func ParseMultiNode(multiEthWSURI string) ([]PeerInfo, error) {
	enodeWsPairs := strings.Split(multiEthWSURI, ",")
	peers := make([]PeerInfo, 0, len(enodeWsPairs))
	for _, pairStr := range enodeWsPairs {
		var peer PeerInfo
		enodeWsPair := strings.Split(pairStr, "+")
		if len(enodeWsPair) == 0 {
			return nil, fmt.Errorf("unable to parse --multi-node. Expected format: enode1+eth-ws-uri-1,enode2+,enode3+eth-ws-uri-3")
		}
		en, err := enode.Parse(enode.ValidSchemes, enodeWsPair[0])
		if err != nil {
			return nil, err
		}
		peer.Enode = en

		ethWsURI := ""
		if len(enodeWsPair) > 1 {
			ethWsURI = enodeWsPair[1]
			if ethWsURI != "" {
				err = validateWsURI(ethWsURI)
				if err != nil {
					return nil, err
				}
			}
		}
		peer.EthWsURI = ethWsURI
		peers = append(peers, peer)
	}
	return peers, nil
}

func validateWsURI(ethWsURI string) error {
	if !strings.HasPrefix(ethWsURI, "ws://") && !strings.HasPrefix(ethWsURI, "wss://") {
		return fmt.Errorf("unable to parse websockets URI [%v]. Expected format: ws(s)://IP:PORT", ethWsURI)
	}
	return nil
}

// StaticEnodes makes a list of enodes only from StaticPeers
func (ec *EthConfig) StaticEnodes() []*enode.Node {
	enodes := make([]*enode.Node, 0, len(ec.StaticPeers))
	for _, peer := range ec.StaticPeers {
		enodes = append(enodes, peer.Enode)
	}
	return enodes
}
