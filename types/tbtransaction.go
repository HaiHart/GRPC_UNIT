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
	txIDs      TxIDList
	addTime    time.Time
	networkNum NetworkNum
}

// NewTbTransaction creates a new transaction to be stored. Transactions are not expected to be initialized with content or txIDs; they should be added via AddTxID and SetContent.
func NewTbTransaction(hash SHA256Hash, networkNum NetworkNum, timestamp time.Time) *TbTransaction {
	return &TbTransaction{
		hash:       hash,
		addTime:    timestamp,
		networkNum: networkNum,
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

// Content returns the transaction contents (usually the blockchain transaction bytes)
func (bt *TbTransaction) Content() TxContent {
	return bt.content
}

// HasContent indicates if transaction has content bytes
func (bt *TbTransaction) HasContent() bool {
	return len(bt.content) > 0
}

// ShortIDs returns the (possibly multiple) short IDs assigned to a transaction
func (bt *TbTransaction) ShortIDs() TxIDList {
	return bt.txIDs
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

// AddTxID adds an assigned shortID, indicating whether it was actually new. Should be called with Lock()
func (bt *TbTransaction) AddTxID(txID TxID) bool {
	if txID == TxIDEmpty {
		return false
	}

	for _, existingShortID := range bt.txIDs {
		if txID == existingShortID {
			return false
		}
	}
	bt.txIDs = append(bt.txIDs, txID)
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
	return EthTransactionFromBytes(bt.hash, bt.content, extractSender)
}
