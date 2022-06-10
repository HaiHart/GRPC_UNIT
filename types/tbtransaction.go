package types

import (
	"sync"
	"time"
)

// TxContent represents a byte array containing full transaction bytes
type TxContent []byte

// TbTransaction represents a single Turbo transaction
type TbTransaction struct {
	m          sync.Mutex
	hash       SHA256Hash
	content    TxContent
	shortIDs   ShortIDList
	addTime    time.Time
	flags      TxFlags
	networkNum NetworkNum
}

// NewTbTransaction creates a new transaction to be stored. Transactions are not expected to be initialized with content or shortIDs; they should be added via AddShortID and SetContent.
func NewTbTransaction(hash SHA256Hash, networkNum NetworkNum, flags TxFlags, timestamp time.Time) *TbTransaction {
	return &TbTransaction{
		hash:       hash,
		addTime:    timestamp,
		networkNum: networkNum,
		flags:      flags,
	}
}

// NewRawTbTransaction creates a new transaction directly from the hash and content. In general, NewRawTbTransaction should not be added directly to TxStore, and should only be validated further before storing.
func NewRawTbTransaction(hash SHA256Hash, content TxContent) *TbTransaction {
	return &TbTransaction{
		hash:    hash,
		content: content,
	}
}

// Hash returns the transaction hash
func (bt *TbTransaction) Hash() SHA256Hash {
	return bt.hash
}

// Flags returns the transaction flags for routing
func (bt *TbTransaction) Flags() TxFlags {
	return bt.flags
}

// AddFlags adds the provided flag to the transaction flag set
func (bt *TbTransaction) AddFlags(flags TxFlags) {
	bt.flags |= flags
}

// SetFlags sets the message flags
func (bt *TbTransaction) SetFlags(flags TxFlags) {
	bt.flags = flags
}

// RemoveFlags sets off txFlag
func (bt *TbTransaction) RemoveFlags(flags TxFlags) {
	bt.SetFlags(bt.Flags() &^ flags)
}

// Content returns the transaction contents (usually the blockchain transaction bytes)
func (bt *TbTransaction) Content() TxContent {
	return bt.content
}

// HasContent indicates if transaction has content bytes
func (bt *TbTransaction) HasContent() bool {
	return len(bt.content) > 0
}

// ShortIDs returns the (possibly multiple) short IDs assigned to a transaction
func (bt *TbTransaction) ShortIDs() ShortIDList {
	return bt.shortIDs
}

// NetworkNum provides the network number of the transaction
func (bt *TbTransaction) NetworkNum() NetworkNum {
	return bt.networkNum
}

// AddTime returns the time the transaction was added
func (bt *TbTransaction) AddTime() time.Time {
	return bt.addTime
}

// SetAddTime sets the time the transaction was added. Should be called with Lock()
func (bt *TbTransaction) SetAddTime(t time.Time) {
	bt.addTime = t
}

// Lock locks the transaction so changes can be made
func (bt *TbTransaction) Lock() {
	bt.m.Lock()
}

// Unlock unlocks the transaction
func (bt *TbTransaction) Unlock() {
	bt.m.Unlock()
}

// AddShortID adds an assigned shortID, indicating whether it was actually new. Should be called with Lock()
func (bt *TbTransaction) AddShortID(shortID ShortID) bool {
	if shortID == ShortIDEmpty {
		return false
	}

	for _, existingShortID := range bt.shortIDs {
		if shortID == existingShortID {
			return false
		}
	}
	bt.shortIDs = append(bt.shortIDs, shortID)
	return true
}

// SetContent sets the blockchain transaction contents only if the contents are new and has never been set before. SetContent returns whether the content was updated. Should be called with Lock()
func (bt *TbTransaction) SetContent(content TxContent) bool {
	if len(bt.content) == 0 && len(content) > 0 {
		bt.content = make(TxContent, len(content))
		copy(bt.content, content)
		return true
	}
	return false
}

// BlockchainTransaction parses and returns a transaction for the given network number's spec
func (bt *TbTransaction) BlockchainTransaction(extractSender bool) (BlockchainTransaction, error) {
	return bt.parseTransaction(extractSender)
}

func (bt *TbTransaction) parseTransaction(extractSender bool) (BlockchainTransaction, error) {
	// TODO - add support for additional networks

	// for now, since we only support Ethereum based transaction
	// we are not checking but parsing as if the transaction is Ethereum based.
	return EthTransactionFromBytes(bt.hash, bt.content, extractSender)
	/*
		switch bt.networkNum {
		case EthereumNetworkNum:
			return NewEthTransaction(bt.hash, bt.content)
		default:
			return nil, fmt.Errorf("no message converter found for network num %v", bt.networkNum)
		}
	*/
}
