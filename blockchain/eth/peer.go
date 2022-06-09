package eth

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	cmap "github.com/orcaman/concurrent-map"
	"math/big"
	"math/rand"
	"strconv"
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

	requestConfirmations bool
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
		requestConfirmations: true,
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

// Disconnect closes the running peer with a protocol error
func (ep *Peer) Disconnect(reason p2p.DiscReason) {
	ep.p.Disconnect(reason)
	ep.disconnected = true
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

// should only be called by blockLoop when then length of the queue has already been checked
func (ep *Peer) sendTopBlock() {
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

// NotifyResponse informs any listeners dependent on a request/response call to this peer, indicating if any channels were waiting for the message
func (ep *Peer) NotifyResponse(packet eth.Packet) bool {
	responseCh := <-ep.responseQueue
	if responseCh != nil {
		responseCh <- packet
	}
	return responseCh != nil
}

// NotifyResponse66 informs any listeners dependent on a request/response call to this ETH66 peer, indicating if any channels were waiting for the message
func (ep *Peer) NotifyResponse66(requestID uint64, packet eth.Packet) (bool, error) {
	rawResponseCh, ok := ep.responseQueue66.Pop(convertRequestIDKey(requestID))
	if !ok {
		return false, ErrUnknownRequestID
	}

	responseCh := rawResponseCh.(chan eth.Packet)
	if responseCh != nil {
		responseCh <- packet
	}
	return responseCh != nil, nil
}

// UpdateHead sets the latest confirmed block on the peer. This may release or prune queued blocks on the peer connection.
func (ep *Peer) UpdateHead(height uint64, hash common.Hash) {
	ep.Log().Debugf("confirming new head (height=%v, hash=%v)", height, hash)
	ep.newHeadCh <- blockRef{
		height: height,
		hash:   hash,
	}
}

// QueueNewBlock adds a new block to the queue to be sent to the peer in the order the peer is ready for.
func (ep *Peer) QueueNewBlock(block *ethtypes.Block, td *big.Int) {

}

// AnnounceBlock pushes a new block announcement to the peer. This is used when the total difficult is unknown, and so a new block message would be invalid.
func (ep *Peer) AnnounceBlock(hash common.Hash, number uint64) error {
	return nil
}

// SendBlockBodies sends a batch of block bodies to the peer
func (ep *Peer) SendBlockBodies(bodies []*eth.BlockBody) error {
	return ep.send(eth.BlockBodiesMsg, eth.BlockBodiesPacket(bodies))
}

// ReplyBlockBodies sends a batch of requested block bodies to the peer (ETH66)
func (ep *Peer) ReplyBlockBodies(id uint64, bodies []*eth.BlockBody) error {
	return ep.send(eth.BlockBodiesMsg, eth.BlockBodiesPacket66{
		RequestId:         id,
		BlockBodiesPacket: bodies,
	})
}

// SendBlockHeaders sends a batch of block headers to the peer
func (ep *Peer) SendBlockHeaders(headers []*ethtypes.Header) error {
	return ep.send(eth.BlockHeadersMsg, eth.BlockHeadersPacket(headers))
}

// ReplyBlockHeaders sends batch of requested block headers to the peer (ETH66)
func (ep *Peer) ReplyBlockHeaders(id uint64, headers []*ethtypes.Header) error {
	return ep.send(eth.BlockHeadersMsg, eth.BlockHeadersPacket66{
		RequestId:          id,
		BlockHeadersPacket: headers,
	})
}

// SendTransactions sends pushes a batch of transactions to the peer
func (ep *Peer) SendTransactions(txs ethtypes.Transactions) error {
	return ep.send(eth.TransactionsMsg, txs)
}

// RequestTransactions requests a batch of announced transactions from the peer
func (ep *Peer) RequestTransactions(txHashes []common.Hash) error {
	return nil
}

// RequestBlock fetches the specified block from the peer, pushing the components to the channels upon request/response completion.
func (ep *Peer) RequestBlock(blockHash common.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet) error {
	if ep.isVersion66() {
		panic("unexpected call to request block for a ETH66 peer")
	}

	getHeadersPacket := &eth.GetBlockHeadersPacket{
		Origin:  eth.HashOrNumber{Hash: blockHash},
		Amount:  1,
		Skip:    0,
		Reverse: false,
	}

	getBodiesPacket := eth.GetBlockBodiesPacket{blockHash}

	ep.registerForResponse(headersCh)
	ep.registerForResponse(bodiesCh)

	if err := ep.send(eth.GetBlockHeadersMsg, getHeadersPacket); err != nil {
		return err
	}
	if err := ep.send(eth.GetBlockBodiesMsg, getBodiesPacket); err != nil {
		return err
	}
	return nil
}

// RequestBlock66 fetches the specified block from the ETH66 peer, pushing the components to the channels upon request/response completion.
func (ep *Peer) RequestBlock66(blockHash common.Hash, headersCh chan eth.Packet, bodiesCh chan eth.Packet) error {
	if !ep.isVersion66() {
		panic("unexpected call to request block 66 for a <ETH66 peer")
	}

	getHeadersPacket := &eth.GetBlockHeadersPacket{
		Origin:  eth.HashOrNumber{Hash: blockHash},
		Amount:  1,
		Skip:    0,
		Reverse: false,
	}
	getBodiesPacket := eth.GetBlockBodiesPacket{blockHash}

	getHeadersRequestID := rand.Uint64()
	getHeadersPacket66 := eth.GetBlockHeadersPacket66{
		RequestId:             getHeadersRequestID,
		GetBlockHeadersPacket: getHeadersPacket,
	}
	getBodiesRequestID := rand.Uint64()
	getBodiesPacket66 := eth.GetBlockBodiesPacket66{
		RequestId:            getBodiesRequestID,
		GetBlockBodiesPacket: getBodiesPacket,
	}

	ep.registerForResponse66(getHeadersRequestID, headersCh)
	ep.registerForResponse66(getBodiesRequestID, headersCh)

	if err := ep.send(eth.GetBlockHeadersMsg, getHeadersPacket66); err != nil {
		return err
	}
	if err := ep.send(eth.GetBlockBodiesMsg, getBodiesPacket66); err != nil {
		return err
	}
	return nil
}

// RequestBlockHeader dispatches a request for the specified block header to the Ethereum peer with no expected handling on response
func (ep *Peer) RequestBlockHeader(hash common.Hash) error {
	return nil
}

// RequestBlockHeaderRaw dispatches a generalized GetHeaders message to the Ethereum peer
func (ep *Peer) RequestBlockHeaderRaw(origin eth.HashOrNumber, amount, skip uint64, reverse bool, responseCh chan eth.Packet) error {
	return nil
}

func (ep *Peer) sendNewBlock(packet *eth.NewBlockPacket) error {
	return nil
}

func (ep *Peer) registerForResponse(responseCh chan eth.Packet) {
	ep.responseQueue <- responseCh
}

func (ep *Peer) registerForResponse66(requestID uint64, responseCh chan eth.Packet) {
	ep.responseQueue66.Set(convertRequestIDKey(requestID), responseCh)
}

func (ep *Peer) send(msgCode uint64, data interface{}) error {
	if ep.disconnected {
		return nil
	}
	return p2p.Send(ep.rw, msgCode, data)
}

func convertRequestIDKey(requestID uint64) string {
	return strconv.FormatUint(requestID, 10)
}
