package services

import (
	"encoding/hex"
	"fmt"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	cmap "github.com/orcaman/concurrent-map"
	"runtime/debug"
	"time"
)

// TbTxStore represents transaction storage of a node
type TbTxStore struct {
	clock         utils.Clock
	hashToContent cmap.ConcurrentMap
	txIDToHash    cmap.ConcurrentMap

	seenTxs            HashHistory
	timeToAvoidReEntry time.Duration

	maxTxAge time.Duration
	assigner TxIDAssigner
}

var _ TxStore = (*TbTxStore)(nil)

func NewTbTxStore(maxTxAge time.Duration, assigner TxIDAssigner, seenTxs HashHistory, timeToAvoidReEntry time.Duration) TbTxStore {
	return TbTxStore{
		clock:              utils.RealClock{},
		hashToContent:      cmap.New(),
		txIDToHash:         cmap.New(),
		seenTxs:            seenTxs,
		timeToAvoidReEntry: timeToAvoidReEntry,
		maxTxAge:           maxTxAge,
		assigner:           assigner,
	}
}

func (t *TbTxStore) Add(hash types.SHA256Hash, content types.TxContent, txID types.TxID, networkNum types.NetworkNum, timestamp time.Time) TransactionResult {
	if txID == types.TxIDEmpty && len(content) == 0 {
		debug.PrintStack()
		panic("content and txID can't be both missing")
	}
	result := TransactionResult{}
	if t.clock.Now().Sub(timestamp) > t.maxTxAge {
		result.Transaction = types.NewTbTransaction(hash, networkNum, timestamp)
		result.DebugData = fmt.Sprintf("Transaction is too old - %v", timestamp)
		return result
	}

	hashStr := string(hash[:])
	// if the hash is in history we treat it as IgnoreSeen
	if t.refreshSeenTx(hash) {
		// if the hash is in history, but we get a txID for it, it means that the hash was not in the ATR history
		// and some GWs may get and use this txID. In such a case we should remove the hash from history and allow
		// it to be added to the TxStore
		if txID != types.TxIDEmpty {
			t.seenTxs.Remove(hashStr)
		} else {
			result.Transaction = types.NewTbTransaction(hash, networkNum, timestamp)
			result.DebugData = fmt.Sprintf("Transaction already seen and deleted from store")
			result.AlreadySeen = true
			return result
		}
	}

	tbTransaction := types.NewTbTransaction(hash, networkNum, timestamp)
	if result.NewTx = t.hashToContent.SetIfAbsent(hashStr, tbTransaction); !result.NewTx {
		tx, exists := t.hashToContent.Get(hashStr)
		if !exists {
			log.Warnf("couldn't Get an existing transaction %v, network %v, txID %v, content %v",
				hash, networkNum, txID, hex.EncodeToString(content[:]))
			result.Transaction = tbTransaction
			result.DebugData = fmt.Sprintf("Transaction deleted by other GO routine")
			return result
		}
		tbTransaction = tx.(*types.TbTransaction)
	}

	// make sure we are the only process that makes changes to the transaction
	tbTransaction.Lock()

	// if txID was not provided, assign txID (if we are running as assigner)
	// note that assigner.Next() provides TxIDEmpty if we are not assigning
	// also, shortID is not assigned if transaction is validators_only
	// if we assigned shortID, result.AssignedTxID hold non ShortIDEmpty value
	if result.NewTx && txID == types.TxIDEmpty {
		txID = t.assigner.Next()
		result.AssignedTxID = txID
	}

	result.NewTxID = tbTransaction.AddTxID(txID)
	result.NewContent = tbTransaction.SetContent(content)
	result.Transaction = tbTransaction
	if result.NewTxID {
		t.txIDToHash.Set(fmt.Sprint(txID), tbTransaction.Hash())
	}
	tbTransaction.Unlock()

	return result
}

func (t *TbTxStore) refreshSeenTx(hash types.SHA256Hash) bool {
	if t.seenTxs.Exists(string(hash[:])) {
		t.seenTxs.Add(string(hash[:]), t.timeToAvoidReEntry)
		return true
	}
	return false
}

// Get returns a single transaction from the transaction service
func (t *TbTxStore) Get(hash types.SHA256Hash) (*types.TbTransaction, bool) {
	// reset the timestamp of this hash in the seenTx hashHistory, if it exists in the hashHistory
	if t.refreshSeenTx(hash) {
		return nil, false
	}
	tx, ok := t.hashToContent.Get(string(hash[:]))
	if !ok {
		return nil, ok
	}
	return tx.(*types.TbTransaction), ok
}

// Known returns whether if a tx hash is in seenTx
func (t *TbTxStore) Known(hash types.SHA256Hash) bool {
	return t.refreshSeenTx(hash)
}
