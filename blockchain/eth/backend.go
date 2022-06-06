package eth

import (
	"context"
	"fmt"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/massbitprotocol/turbo/blockchain/network"
	log "github.com/massbitprotocol/turbo/logger"
	"time"
)

const (
	checkpointTimeout    = 5 * time.Second
	maxFutureBlockNumber = 100
)

// Backend represents the interface to which any stateful message handling (e.g. looking up tx pool items or block headers) will be passed to for processing
type Backend interface {
	NetworkConfig() *network.EthConfig
	RunPeer(peer *Peer, handler func(*Peer) error) error
	Handle(peer *Peer, packet eth.Packet) error
	GetHeaders(start eth.HashOrNumber, count int, skip int, reverse bool) ([]*ethtypes.Header, error)
	GetBodies(hashes []ethcommon.Hash) ([]*ethtypes.Body, error)
}

// Handler is the Ethereum backend implementation. It tracks received blocks and transactions from peers and passes transactions and blocks to the bridge.
type Handler struct {
	peers  *peerSet
	cancel context.CancelFunc
	config *network.EthConfig

	chain *Chain
}

// NewHandler returns a new Handler and starts its processing goroutines
func NewHandler(parent context.Context, config *network.EthConfig) *Handler {
	ctx, cancel := context.WithCancel(parent)
	h := &Handler{
		peers:  newPeerSet(),
		cancel: cancel,
		config: config,
		chain:  NewChain(ctx),
	}
	return h
}

// NetworkConfig returns the backend's network configuration
func (h *Handler) NetworkConfig() *network.EthConfig {
	return h.config
}

// RunPeer registers a peer within the peer set and starts handling all its messages
func (h *Handler) RunPeer(ep *Peer, handler func(*Peer) error) error {
	if err := h.peers.register(ep); err != nil {
		return err
	}
	defer func() {
		ep.ctx.Done()
		ep.Stop()
		_ = h.peers.unregister(ep.ID())
	}()

	ep.Start()
	time.AfterFunc(checkpointTimeout, func() {
		ep.checkpointPassed = true
	})
	return handler(ep)
}

// Handle processes Ethereum message packets that update internal backend state. In general, messages that require responses should not reach this function.
func (h *Handler) Handle(peer *Peer, packet eth.Packet) error {
	switch p := packet.(type) {
	case *eth.StatusPacket:
		h.chain.InitializeDifficulty(p.Head, p.TD)
		return nil
	case *eth.TransactionsPacket:
		return h.processTransactions(peer, *p)
	case *eth.PooledTransactionsPacket:
		return h.processTransactions(peer, *p)
	case *eth.NewPooledTransactionHashesPacket:
		return h.processTransactionHashes(peer, *p)
	case *eth.NewBlockPacket:
		return h.processBlock(peer, NewBlockInfo(p.Block, p.TD))
	case *eth.NewBlockHashesPacket:
		return h.processBlockAnnouncement(peer, *p)
	case *eth.BlockHeadersPacket:
		return h.processBlockHeaders(peer, *p)
	default:
		return fmt.Errorf("unexpected eth packet type: %v", packet)
	}
}

func (h *Handler) processTransactions(peer *Peer, txs []*ethtypes.Transaction) error {
	log.Infof("process transactions: %v", txs)
	return nil
}

func (h *Handler) processTransactionHashes(peer *Peer, txHashes []ethcommon.Hash) error {
	log.Infof("process transaction hashes: %v", txHashes)
	return nil
}

func (h *Handler) processBlock(peer *Peer, blockInfo *BlockInfo) error {
	log.Infof("process block: %v", blockInfo)
	return nil
}

func (h *Handler) processBlockAnnouncement(peer *Peer, newBlocks eth.NewBlockHashesPacket) error {
	log.Infof("process block announcement: %v", newBlocks)
	return nil
}

func (h *Handler) processBlockHeaders(peer *Peer, blockHeaders eth.BlockHeadersPacket) error {
	log.Infof("process block headers: %v", blockHeaders)
	return nil
}

// GetHeaders assembles and returns a set of headers
func (h *Handler) GetHeaders(start eth.HashOrNumber, count int, skip int, reverse bool) ([]*ethtypes.Header, error) {
	return h.chain.GetHeaders(start, count, skip, reverse)
}

// GetBodies assembles and returns a set of block bodies
func (h *Handler) GetBodies(hashes []ethcommon.Hash) ([]*ethtypes.Body, error) {
	return h.chain.GetBodies(hashes)
}
