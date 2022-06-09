package eth

import (
	"context"
	"errors"
	"fmt"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/massbitprotocol/turbo/blockchain/network"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"math/big"
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
	chain  *Chain
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

func (h *Handler) awaitBlockResponse(peer *Peer, blockHash ethcommon.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet, fetchResponse func(peer *Peer, blockHash ethcommon.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet) (*eth.BlockHeadersPacket, *eth.BlockBodiesPacket, error)) {
	startTime := time.Now()
	headers, bodies, err := fetchResponse(peer, blockHash, headersCh, bodiesCh)

	if err != nil {
		if peer.disconnected {
			peer.Log().Tracef("block %v response timed out for disconnected peer", blockHash)
			return
		}

		if err == ErrInvalidPacketType {
			// message is already logged
		} else if err == ErrResponseTimeout {
			peer.Log().Errorf("did not receive block header and body for block %v before timeout", blockHash)
		} else {
			peer.Log().Errorf("could not fetch block header and body for block %v: %v", blockHash, err)
		}
		peer.Disconnect(p2p.DiscUselessPeer)
		return
	}

	elapsedTime := time.Now().Sub(startTime)
	peer.Log().Debugf("took %v to fetch block %v header and body", elapsedTime, blockHash)

	if err := h.processBlockComponents(peer, headers, bodies); err != nil {
		log.Errorf("error processing block components for hash %v: %v", blockHash, err)
	}
}

func (h *Handler) fetchBlockResponse66(peer *Peer, blockHash ethcommon.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet) (*eth.BlockHeadersPacket, *eth.BlockBodiesPacket, error) {
	var (
		headers *eth.BlockHeadersPacket
		bodies  *eth.BlockBodiesPacket
		errCh   = make(chan error, 2)
	)

	go func() {
		var ok bool
		select {
		case rawHeaders := <-headersCh:
			headers, ok = rawHeaders.(*eth.BlockHeadersPacket)
			peer.Log().Debugf("received header for block %v", blockHash)
			if !ok {
				log.Errorf("could not convert headers for block %v to the expected packet type, got %T", blockHash, rawHeaders)
				errCh <- ErrInvalidPacketType
			} else {
				errCh <- nil
			}
		case <-time.After(responseTimeout):
			errCh <- ErrResponseTimeout
		}
	}()

	go func() {
		var ok bool

		select {
		case rawBodies := <-bodiesCh:
			bodies, ok = rawBodies.(*eth.BlockBodiesPacket)
			peer.Log().Debugf("received body for block %v", blockHash)
			if !ok {
				log.Errorf("could not convert bodies for block %v to the expected packet type, got %T", blockHash, rawBodies)
				errCh <- ErrInvalidPacketType
			} else {
				errCh <- nil
			}
		case <-time.After(responseTimeout):
			errCh <- ErrResponseTimeout
		}
	}()

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			return nil, nil, err
		}
	}

	return headers, bodies, nil
}

func (h *Handler) fetchBlockResponse(peer *Peer, blockHash ethcommon.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet) (*eth.BlockHeadersPacket, *eth.BlockBodiesPacket, error) {
	var (
		headers *eth.BlockHeadersPacket
		bodies  *eth.BlockBodiesPacket
		ok      bool
	)

	select {
	case rawHeaders := <-headersCh:
		headers, ok = rawHeaders.(*eth.BlockHeadersPacket)
		if !ok {
			log.Errorf("could not convert headers for block %v to the expected packet type, got %T", blockHash, rawHeaders)
			return nil, nil, ErrInvalidPacketType
		}
	case <-time.After(responseTimeout):
		return nil, nil, ErrResponseTimeout
	}

	peer.Log().Debugf("received header for block %v", blockHash)

	select {
	case rawBodies := <-bodiesCh:
		bodies, ok = rawBodies.(*eth.BlockBodiesPacket)
		if !ok {
			log.Errorf("could not convert headers for block %v to the expected packet type, got %T", blockHash, rawBodies)
			return nil, nil, ErrInvalidPacketType
		}
	case <-time.After(responseTimeout):
		return nil, nil, ErrResponseTimeout
	}

	peer.Log().Debugf("received body for block %v", blockHash)

	return headers, bodies, nil
}

func (h *Handler) processBlockComponents(peer *Peer, headers *eth.BlockHeadersPacket, bodies *eth.BlockBodiesPacket) error {
	if len(*headers) != 1 && len(*bodies) != 1 {
		return fmt.Errorf("received %v headers and %v bodies, instead of 1 of each", len(*headers), len(*bodies))
	}

	header := (*headers)[0]
	body := (*bodies)[0]
	block := ethtypes.NewBlockWithHeader(header).WithBody(body.Transactions, body.Uncles)
	blockInfo := NewBlockInfo(block, nil)
	_ = h.chain.SetTotalDifficulty(blockInfo)
	return h.processBlock(peer, blockInfo)
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

func (h *Handler) broadcastBlock(block *ethtypes.Block, totalDifficulty *big.Int, sourceBlockchainPeer *Peer) {

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
	block := blockInfo.Block
	blockHash := block.Hash()
	blockHeight := block.Number()

	if err := h.validateBlock(blockHash, blockHeight.Int64(), int64(block.Time())); err != nil {
		if err == ErrAlreadySeen {
			peer.Log().Debugf("skipping block %v (height %v): %v", blockHash, blockHeight, err)
		} else {
			peer.Log().Warnf("skipping block %v (height %v): %v", blockHash, blockHeight, err)
		}
		return nil
	}

	peer.Log().Debugf("processing new block %v (height %v)", blockHash, blockHeight)
	newHeadCount := h.chain.AddBlock(blockInfo, BSBlockchain)
	h.sendConfirmedBlocksToBDN(newHeadCount, peer.IPEndpoint())
	h.broadcastBlock(block, blockInfo.totalDifficulty, peer)
	return nil
}

func (h *Handler) validateBlock(blockHash ethcommon.Hash, blockHeight int64, blockTime int64) error {
	if err := h.chain.ValidateBlock(blockHash, blockHeight); err != nil {
		return err
	}
	minTimestamp := time.Now().Add(-h.config.IgnoreBlockTimeout)
	if minTimestamp.Unix() > blockTime {
		return errors.New("timestamp too old")
	}
	return nil
}

func (h *Handler) processBlockAnnouncement(peer *Peer, newBlocks eth.NewBlockHashesPacket) error {
	for _, block := range newBlocks {
		peer.Log().Debugf("processing new block announcement %v (height %v)", block.Hash, block.Number)

		if !h.chain.HasBlock(block.Hash) {
			headersCh := make(chan eth.Packet)
			bodiesCh := make(chan eth.Packet)

			if peer.isVersion66() {
				err := peer.RequestBlock66(block.Hash, headersCh, bodiesCh)
				if err != nil {
					peer.Log().Errorf("could not request block %v: %v", block.Hash, err)
					return err
				}
				go h.awaitBlockResponse(peer, block.Hash, headersCh, bodiesCh, h.fetchBlockResponse66)
			} else {
				err := peer.RequestBlock(block.Hash, headersCh, bodiesCh)
				if err != nil {
					peer.Log().Errorf("could not request block %v: %v", block.Hash, err)
					return err
				}
				go h.awaitBlockResponse(peer, block.Hash, headersCh, bodiesCh, h.fetchBlockResponse)
			}
		} else {
			h.confirmBlock(block.Hash, peer.IPEndpoint())
		}
	}

	return nil
}

func (h *Handler) processBlockHeaders(peer *Peer, blockHeaders eth.BlockHeadersPacket) error {
	return nil
}

func (h *Handler) confirmBlock(hash ethcommon.Hash, peerEndpoint types.NodeEndpoint) {

}

func (h *Handler) sendConfirmedBlocksToBDN(count int, peerEndpoint types.NodeEndpoint) {

}

// GetHeaders assembles and returns a set of headers
func (h *Handler) GetHeaders(start eth.HashOrNumber, count int, skip int, reverse bool) ([]*ethtypes.Header, error) {
	return h.chain.GetHeaders(start, count, skip, reverse)
}

// GetBodies assembles and returns a set of block bodies
func (h *Handler) GetBodies(hashes []ethcommon.Hash) ([]*ethtypes.Body, error) {
	return h.chain.GetBodies(hashes)
}
