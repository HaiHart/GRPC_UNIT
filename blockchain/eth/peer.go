package eth

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	cmap "github.com/orcaman/concurrent-map"
	"math/big"
	"time"
)

const (
	maxMessageSize                  = 10 * 1024 * 1024
	responseQueueSize               = 10
	responseTimeout                 = 5 * time.Minute
	fastBlockConfirmationInterval   = 200 * time.Millisecond
	fastBlockConfirmationAttempts   = 5
	slowBlockConfirmationInterval   = 1 * time.Second
	headChannelBacklog              = 10
	blockChannelBacklog             = 10
	blockConfirmationChannelBacklog = 10
	blockQueueMaxSize               = 50
)

// special error constants during peer message processing
var (
	ErrResponseTimeout  = errors.New("response timed out")
	ErrUnknownRequestID = errors.New("unknown request ID on message")
)

// Peer wraps an Ethereum peer
type Peer struct {
	p        *p2p.Peer
	rw       p2p.MsgReadWriter
	version  uint
	endpoint types.NodeEndpoint
	clock    utils.Clock
	ctx      context.Context
	cancel   context.CancelFunc
	log      *log.Entry

	disconnected     bool
	checkpointPassed bool

	responseQueue   chan chan eth.Packet
	responseQueue66 cmap.ConcurrentMap

	newHeadCh           chan blockRef
	newBlockCh          chan *eth.NewBlockPacket
	blockConfirmationCh chan common.Hash
	confirmedHead       blockRef
	sentHead            blockRef
	queuedBlocks        []*eth.NewBlockPacket

	RequestConfirmations bool
}

// NewPeer returns a wrapped Ethereum peer
func NewPeer(ctx context.Context, p *p2p.Peer, rw p2p.MsgReadWriter, version uint) *Peer {
	return newPeer(ctx, p, rw, version, utils.RealClock{})
}

func newPeer(parent context.Context, p *p2p.Peer, rw p2p.MsgReadWriter, version uint, clock utils.Clock) *Peer {
	ctx, cancel := context.WithCancel(parent)
	peer := &Peer{
		p:       p,
		rw:      rw,
		version: version,
		endpoint: types.NodeEndpoint{
			IP:        p.Node().IP().String(),
			Port:      p.Node().TCP(),
			PublicKey: p.Info().Enode,
		},
		clock:                clock,
		ctx:                  ctx,
		cancel:               cancel,
		responseQueue:        make(chan chan eth.Packet, responseQueueSize),
		responseQueue66:      cmap.New(),
		newHeadCh:            make(chan blockRef, headChannelBacklog),
		newBlockCh:           make(chan *eth.NewBlockPacket, blockChannelBacklog),
		blockConfirmationCh:  make(chan common.Hash, blockConfirmationChannelBacklog),
		queuedBlocks:         make([]*eth.NewBlockPacket, 0),
		RequestConfirmations: true,
	}
	peerID := p.ID()
	peer.log = log.WithFields(log.Fields{
		"connType":   "ETH",
		"remoteAddr": p.RemoteAddr().String(),
		"id":         fmt.Sprintf("%x", peerID[:8]),
	})
	return peer
}

// ID provides a unique identifier for each Ethereum peer
func (ep *Peer) ID() string {
	return ep.p.ID().String()
}

// String formats the peer for display
func (ep *Peer) String() string {
	id := ep.p.ID()
	return fmt.Sprintf("ETH/%x@%v", id[:8], ep.p.RemoteAddr())
}

// IPEndpoint provides the peer IP endpoint
func (ep *Peer) IPEndpoint() types.NodeEndpoint {
	return ep.endpoint
}

// Log returns the context logger for the peer connection
func (ep *Peer) Log() *log.Entry {
	return ep.log.WithField("head", ep.confirmedHead)
}

// Start launches the block sending loop that queued blocks get sent in order to the peer
func (ep *Peer) Start() {
	go ep.blockLoop()
}

// Stop shuts down the running goroutines
func (ep *Peer) Stop() {
	ep.cancel()
}

func (ep *Peer) isVersion66() bool {
	return ep.version >= eth.ETH66
}

func (ep *Peer) blockLoop() {
	for {
		select {}
	}
}

// Handshake executes the Ethereum protocol Handshake. Unlike Geth, the gateway waits for the peer status message before sending its own, in order to replicate some peer status fields.
func (ep *Peer) Handshake(version uint32, network uint64, td *big.Int, head common.Hash, genesis common.Hash) (*eth.StatusPacket, error) {
	peerStatus, err := ep.readStatus()
	if err != nil {
		return peerStatus, err
	}

	ep.confirmedHead = blockRef{hash: head}

	// used same fork ID as received from peer; gateway is expected to usually be compatible with Ethereum peer
	err = ep.send(eth.StatusMsg, &eth.StatusPacket{
		ProtocolVersion: version,
		NetworkID:       network,
		TD:              td,
		Head:            head,
		Genesis:         genesis,
		ForkID:          peerStatus.ForkID,
	})

	if peerStatus.NetworkID != network {
		return peerStatus, fmt.Errorf("network ID does not match: expected %v, but got %v", network, peerStatus.NetworkID)
	}

	if peerStatus.ProtocolVersion != version {
		return peerStatus, fmt.Errorf("protocol version does not match: expected %v, but got %v", version, peerStatus.ProtocolVersion)
	}

	if peerStatus.Genesis != genesis {
		return peerStatus, fmt.Errorf("genesis block does not match: expected %v, but got %v", genesis, peerStatus.Genesis)
	}

	return peerStatus, nil
}

func (ep *Peer) readStatus() (*eth.StatusPacket, error) {
	var status eth.StatusPacket

	msg, err := ep.rw.ReadMsg()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = msg.Discard()
	}()

	if msg.Code != eth.StatusMsg {
		return &status, fmt.Errorf("unexpected first message: %v, should have been %v", msg.Code, eth.StatusMsg)
	}

	if msg.Size > maxMessageSize {
		return &status, fmt.Errorf("message is too big: %v > %v", msg.Size, maxMessageSize)
	}

	if err = msg.Decode(&status); err != nil {
		return &status, fmt.Errorf("could not decode status message: %v", err)
	}

	return &status, nil
}

func (ep *Peer) send(msgCode uint64, data interface{}) error {
	if ep.disconnected {
		return nil
	}
	return p2p.Send(ep.rw, msgCode, data)
}
