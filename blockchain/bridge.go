package blockchain

import (
	"errors"
	"github.com/massbitprotocol/turbo/blockchain/network"
	"github.com/massbitprotocol/turbo/types"
)

// constants for transaction channel buffer sizes
const (
	transactionBacklog       = 500
	transactionHashesBacklog = 1000
	blockBacklog             = 100
	statusBacklog            = 10
)

// ErrChannelFull is a special error for identifying overflowing channel buffers
var ErrChannelFull = errors.New("channel full")

// NoActiveBlockchainPeersAlert is used to send an alert to the gateway on initial liveliness check if no active blockchain peers
type NoActiveBlockchainPeersAlert struct {
}

// TransactionAnnouncement represents an available transaction from a given peer that can be requested
type TransactionAnnouncement struct {
	Hashes types.SHA256HashList
	PeerID string
}

// TransactionsFromNode is used to pass transactions from a node to the BDN
type TransactionsFromNode struct {
	Transactions []*types.TbTransaction
	PeerEndpoint types.NodeEndpoint
}

// BlockFromNode is used to pass blocks from a node to the BDN
type BlockFromNode struct {
	Block        *types.TbBlock
	PeerEndpoint types.NodeEndpoint
}

// BlockAnnouncement represents an available block from a given peer that can be requested
type BlockAnnouncement struct {
	Hash         types.SHA256Hash
	PeerID       string
	PeerEndpoint types.NodeEndpoint
}

// Converter defines an interface for converting between blockchain and BDN transactions
type Converter interface {
	TransactionBlockchainToBDN(interface{}) (*types.TbTransaction, error)
	TransactionBDNToBlockchain(*types.TbTransaction) (interface{}, error)
	BlockBlockchainToBDN(interface{}) (*types.TbBlock, error)
	BlockBDNtoBlockchain(block *types.TbBlock) (interface{}, error)
}

// Bridge represents the application interface over which messages are passed between the blockchain node and the BDN
type Bridge interface {
	Converter

	ReceiveNetworkConfigUpdates() <-chan network.EthConfig
	UpdateNetworkConfig(network.EthConfig) error

	AnnounceTransactionHashes(string, types.SHA256HashList) error
	SendTransactionsFromBDN([]*types.TbTransaction) error
	SendTransactionsToBDN(txs []*types.TbTransaction, peerEndpoint types.NodeEndpoint) error
	RequestTransactionsFromNode(string, types.SHA256HashList) error

	ReceiveNodeTransactions() <-chan TransactionsFromNode
	ReceiveBDNTransactions() <-chan []*types.TbTransaction
	ReceiveTransactionHashesAnnouncement() <-chan TransactionAnnouncement
	ReceiveTransactionHashesRequest() <-chan TransactionAnnouncement

	SendBlockToBDN(*types.TbBlock, types.NodeEndpoint) error
	SendBlockToNode(*types.TbBlock) error
	SendConfirmedBlockToGateway(block *types.TbBlock, peerEndpoint types.NodeEndpoint) error

	ReceiveBlockFromBDN() <-chan *types.TbBlock
	ReceiveBlockFromNode() <-chan BlockFromNode
	ReceiveConfirmedBlockFromNode() <-chan BlockFromNode

	ReceiveNoActiveBlockchainPeersAlert() <-chan NoActiveBlockchainPeersAlert
	SendNoActiveBlockchainPeersAlert() error

	SendBlockchainStatusRequest() error
	ReceiveBlockchainStatusRequest() <-chan struct{}
	SendBlockchainStatusResponse([]*types.NodeEndpoint) error
	ReceiveBlockchainStatusResponse() <-chan []*types.NodeEndpoint
}

// TurboBridge is an implementation of the Bridge interface
type TurboBridge struct {
	Converter
	config                    chan network.EthConfig
	transactionsFromNode      chan TransactionsFromNode
	transactionsFromBDN       chan []*types.TbTransaction
	transactionHashesFromNode chan TransactionAnnouncement
	transactionHashesRequests chan TransactionAnnouncement

	blocksFromNode         chan BlockFromNode
	blocksFromBDN          chan *types.TbBlock
	confirmedBlockFromNode chan BlockFromNode

	noActiveBlockchainPeers chan NoActiveBlockchainPeersAlert

	blockchainStatusRequest  chan struct{}
	blockchainStatusResponse chan []*types.NodeEndpoint
}

// NewTurboBridge returns a TurboBridge instance
func NewTurboBridge(converter Converter) Bridge {
	return &TurboBridge{
		config:                    make(chan network.EthConfig),
		transactionsFromNode:      make(chan TransactionsFromNode, transactionBacklog),
		transactionsFromBDN:       make(chan []*types.TbTransaction, transactionBacklog),
		transactionHashesFromNode: make(chan TransactionAnnouncement, transactionHashesBacklog),
		transactionHashesRequests: make(chan TransactionAnnouncement, transactionHashesBacklog),
		blocksFromNode:            make(chan BlockFromNode, blockBacklog),
		blocksFromBDN:             make(chan *types.TbBlock, blockBacklog),
		confirmedBlockFromNode:    make(chan BlockFromNode, blockBacklog),
		noActiveBlockchainPeers:   make(chan NoActiveBlockchainPeersAlert),
		blockchainStatusRequest:   make(chan struct{}, statusBacklog),
		blockchainStatusResponse:  make(chan []*types.NodeEndpoint, statusBacklog),
		Converter:                 converter,
	}
}

// ReceiveNetworkConfigUpdates provides a channel with network config updates
func (b *TurboBridge) ReceiveNetworkConfigUpdates() <-chan network.EthConfig {
	return b.config
}

// UpdateNetworkConfig pushes a new Ethereum configuration update
func (b *TurboBridge) UpdateNetworkConfig(config network.EthConfig) error {
	b.config <- config
	return nil
}

// AnnounceTransactionHashes pushes a series of transaction announcements onto the announcements channel
func (b TurboBridge) AnnounceTransactionHashes(peerID string, hashes types.SHA256HashList) error {
	select {
	case b.transactionHashesFromNode <- TransactionAnnouncement{Hashes: hashes, PeerID: peerID}:
		return nil
	default:
		return ErrChannelFull
	}
}

// RequestTransactionsFromNode requests a series of transactions that a peer node has announced
func (b TurboBridge) RequestTransactionsFromNode(peerID string, hashes types.SHA256HashList) error {
	select {
	case b.transactionHashesRequests <- TransactionAnnouncement{Hashes: hashes, PeerID: peerID}:
		return nil
	default:
		return ErrChannelFull
	}
}

// SendTransactionsFromBDN sends a set of transactions from the BDN for distribution to nodes
func (b TurboBridge) SendTransactionsFromBDN(transactions []*types.TbTransaction) error {
	select {
	case b.transactionsFromBDN <- transactions:
		return nil
	default:
		return ErrChannelFull
	}
}

// SendTransactionsToBDN sends a set of transactions from a node to the BDN for propagation
func (b TurboBridge) SendTransactionsToBDN(txs []*types.TbTransaction, peerEndpoint types.NodeEndpoint) error {
	select {
	case b.transactionsFromNode <- TransactionsFromNode{Transactions: txs, PeerEndpoint: peerEndpoint}:
		return nil
	default:
		return ErrChannelFull
	}
}

// SendConfirmedBlockToGateway sends a SHA256 of the block to be included in blockConfirm message
func (b TurboBridge) SendConfirmedBlockToGateway(block *types.TbBlock, peerEndpoint types.NodeEndpoint) error {
	select {
	case b.confirmedBlockFromNode <- BlockFromNode{Block: block, PeerEndpoint: peerEndpoint}:
		return nil
	default:
		return ErrChannelFull
	}
}

// ReceiveNodeTransactions provides a channel that pushes transactions as they come in from nodes
func (b TurboBridge) ReceiveNodeTransactions() <-chan TransactionsFromNode {
	return b.transactionsFromNode
}

// ReceiveBDNTransactions provides a channel that pushes transactions as they arrive from the BDN
func (b TurboBridge) ReceiveBDNTransactions() <-chan []*types.TbTransaction {
	return b.transactionsFromBDN
}

// ReceiveTransactionHashesAnnouncement provides a channel that pushes announcements as nodes announce them
func (b TurboBridge) ReceiveTransactionHashesAnnouncement() <-chan TransactionAnnouncement {
	return b.transactionHashesFromNode
}

// ReceiveTransactionHashesRequest provides a channel that pushes requests for transaction hashes from the BDN
func (b TurboBridge) ReceiveTransactionHashesRequest() <-chan TransactionAnnouncement {
	return b.transactionHashesRequests
}

// SendBlockToBDN sends a block from a node to the BDN
func (b TurboBridge) SendBlockToBDN(block *types.TbBlock, peerEndpoint types.NodeEndpoint) error {
	select {
	case b.blocksFromNode <- BlockFromNode{Block: block, PeerEndpoint: peerEndpoint}:
		return nil
	default:
		return ErrChannelFull
	}
}

// SendBlockToNode sends a block from the BDN for distribution to nodes
func (b TurboBridge) SendBlockToNode(block *types.TbBlock) error {
	select {
	case b.blocksFromBDN <- block:
		return nil
	default:
		return ErrChannelFull
	}
}

// ReceiveBlockFromNode provides a channel that pushes blocks as they come in from nodes
func (b TurboBridge) ReceiveBlockFromNode() <-chan BlockFromNode {
	return b.blocksFromNode
}

// ReceiveBlockFromBDN provides a channel that pushes new blocks from the BDN
func (b TurboBridge) ReceiveBlockFromBDN() <-chan *types.TbBlock {
	return b.blocksFromBDN
}

// ReceiveConfirmedBlockFromNode provides a channel that pushes confirmed blocks from nodes
func (b TurboBridge) ReceiveConfirmedBlockFromNode() <-chan BlockFromNode {
	return b.confirmedBlockFromNode
}

// SendNoActiveBlockchainPeersAlert sends alerts to the BDN when there is no active blockchain peer
func (b TurboBridge) SendNoActiveBlockchainPeersAlert() error {
	select {
	case b.noActiveBlockchainPeers <- NoActiveBlockchainPeersAlert{}:
		return nil
	default:
		return ErrChannelFull
	}
}

// ReceiveNoActiveBlockchainPeersAlert provides a channel that pushes no active blockchain peer alerts
func (b TurboBridge) ReceiveNoActiveBlockchainPeersAlert() <-chan NoActiveBlockchainPeersAlert {
	return b.noActiveBlockchainPeers
}

// SendBlockchainStatusRequest sends a blockchain connection status request signal to a blockchain backend
func (b TurboBridge) SendBlockchainStatusRequest() error {
	select {
	case b.blockchainStatusRequest <- struct{}{}:
		return nil
	default:
		return ErrChannelFull
	}
}

// ReceiveBlockchainStatusRequest handles SendBlockchainStatusRequest signal
func (b TurboBridge) ReceiveBlockchainStatusRequest() <-chan struct{} {
	return b.blockchainStatusRequest
}

// SendBlockchainStatusResponse sends a response for blockchain connection status request
func (b TurboBridge) SendBlockchainStatusResponse(endpoints []*types.NodeEndpoint) error {
	select {
	case b.blockchainStatusResponse <- endpoints:
		return nil
	default:
		return ErrChannelFull
	}
}

// ReceiveBlockchainStatusResponse handles blockchain connection status response from backend
func (b TurboBridge) ReceiveBlockchainStatusResponse() <-chan []*types.NodeEndpoint {
	return b.blockchainStatusResponse
}
