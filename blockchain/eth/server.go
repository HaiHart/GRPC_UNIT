package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/massbitprotocol/turbo/blockchain/network"
	"path"
)

// Server wraps the Ethereum p2p server
type Server struct {
	*p2p.Server
	cancel context.CancelFunc
}

func NewServer(parent context.Context, config *network.EthConfig, dataDir string, logger log.Logger) (*Server, error) {
	var privateKey *ecdsa.PrivateKey
	if config.PrivateKey != nil {
		privateKey = config.PrivateKey
	} else {
		privateKeyPath := path.Join(dataDir, ".gatewaykey")
		privateKeyFromFile, generated, err := LoadOrGeneratePrivateKey(privateKeyPath)

		if err != nil {
			keyWriteErr, ok := err.(keyWriteError)
			if ok {
				logger.Warn("could not write private key", "err", keyWriteErr)
			} else {
				return nil, err
			}
		}

		if generated {
			logger.Warn("no private key found, generating new one", "path", privateKeyPath)
		}
		privateKey = privateKeyFromFile
	}

	ctx, cancel := context.WithCancel(parent)
	backend := NewHandler(ctx, config)

	server := p2p.Server{
		Config: p2p.Config{
			PrivateKey:       privateKey,
			MaxPeers:         len(config.StaticPeers),
			MaxPendingPeers:  1,
			DialRatio:        1,
			NoDiscovery:      true,
			DiscoveryV5:      false,
			Name:             fmt.Sprintf("Turbo Gateway v0.1.0"),
			BootstrapNodes:   nil,
			BootstrapNodesV5: nil,
			StaticNodes:      config.StaticEnodes(),
			TrustedNodes:     nil,
			NetRestrict:      nil,
			NodeDatabase:     "",
			Protocols:        MakeProtocols(ctx, backend),
			ListenAddr:       "",
			NAT:              nil,
			Dialer:           nil,
			NoDial:           false,
			EnableMsgEvents:  false,
			Logger:           logger,
		},
		DiscV5: nil,
	}

	s := &Server{
		Server: &server,
		cancel: cancel,
	}
	return s, nil
}
