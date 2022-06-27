package services

import (
	"encoding/hex"
	"fmt"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	cmap "github.com/orcaman/concurrent-map"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

// TxStoreMaxSize - If number of Txs in TxStore is above TxStoreMaxSize cleanup will bring it back to TxStoreMaxSize (per network)
const TxStoreMaxSize = 200000

// TbTxStore represents transaction storage of a node
type TbTxStore struct {
	clock         utils.Clock
	hashToContent cmap.ConcurrentMap
	txIDToHash    cmap.ConcurrentMap

	seenTxs            HashHistory
	timeToAvoidReEntry time.Duration

	cleanupFreq         time.Duration
	maxTxAge            time.Duration
	noTxIDAge           time.Duration
	quit                chan bool
	lock                sync.Mutex
	assigner            TxIDAssigner
	cleanedTxIDsChannel chan types.TxIDsByNetwork
}

var _ TxStore = (*TbTxStore)(nil)

func NewTbTxStore(cleanupFreq time.Duration, maxTxAge time.Duration, noTxIDAge time.Duration, assigner TxIDAssigner, seenTxs HashHistory, cleanedTxIDsChannel chan types.TxIDsByNetwork, timeToAvoidReEntry time.Duration) TbTxStore {
	return TbTxStore{
		clock:               utils.RealClock{},
		hashToContent:       cmap.New(),
		txIDToHash:          cmap.New(),
		seenTxs:             seenTxs,
		timeToAvoidReEntry:  timeToAvoidReEntry,
		cleanupFreq:         cleanupFreq,
		maxTxAge:            maxTxAge,
		noTxIDAge:           noTxIDAge,
		quit:                make(chan bool),
		assigner:            assigner,
		cleanedTxIDsChannel: cleanedTxIDsChannel,
	}
}

// Start initializes all relevant goroutines for the TbTxStore
func (t *TbTxStore) Start() error {
	t.cleanup()
	return nil
}

// Stop closes all running go routines for TbTxStore
func (t *TbTxStore) Stop() {
	t.quit <- true
	<-t.quit
}

// Clear removes all elements from txs and txIDToHash
func (t *TbTxStore) Clear() {
	t.hashToContent.Clear()
	t.txIDToHash.Clear()
	log.Debugf("Cleared tx service.")
}

// Count indicates the number of stored transaction in TbTxStore
func (t *TbTxStore) Count() int {
	return t.hashToContent.Count()
}

// remove deletes a single transaction, including its txIDs
func (t *TbTxStore) remove(hash string, reEntryProtection ReEntryProtectionFlags, reason string) {
	if tx, ok := t.hashToContent.Pop(hash); ok {
		tbTransaction := tx.(*types.TbTransaction)
		for _, txID := range tbTransaction.TxIDs() {
			t.txIDToHash.Remove(fmt.Sprint(txID))
		}
		// if asked, add the hash to the history map, so we remember this transaction for some time
		// and prevent if from being added back to the TxStore
		switch reEntryProtection {
		case NoReEntryProtection:
		case ShortReEntryProtection:
			t.seenTxs.Add(hash, ShortReEntryProtectionDuration)
		case FullReEntryProtection:
			t.seenTxs.Add(hash, t.timeToAvoidReEntry)
		default:
			log.Fatalf("unknown reEntryProtection value %v for hash %v", reEntryProtection, hash)
		}
		log.Tracef("TxStore: transaction %v, network %v, shortIDs %v removed (%v). reEntryProtection %v",
			tbTransaction.Hash(), tbTransaction.NetworkNum(), tbTransaction.TxIDs(), reason, reEntryProtection)
	}
}

// RemoveTxIDs deletes a series of transactions by their short IDs. RemoveTxIDs can take a potentially large ID array, so it should be passed by reference.
func (t *TbTxStore) RemoveTxIDs(txIDs *types.TxIDList, reEntryProtection ReEntryProtectionFlags, reason string) {
	// note - it is OK for hashesToRemove to hold the same hash multiple times.
	hashesToRemove := make(types.SHA256HashList, 0)
	for _, txID := range *txIDs {
		if hash, ok := t.txIDToHash.Get(fmt.Sprint(txID)); ok {
			hashesToRemove = append(hashesToRemove, hash.(types.SHA256Hash))
		}
	}
	t.RemoveHashes(&hashesToRemove, reEntryProtection, reason)
}

// GetTxByID lookup a transaction by its txID. return error if not found
func (t *TbTxStore) GetTxByID(txID types.TxID) (*types.TbTransaction, error) {
	if h, ok := t.txIDToHash.Get(fmt.Sprint(txID)); ok {
		hash := h.(types.SHA256Hash)
		if tx, ok := t.hashToContent.Get(string(hash[:])); ok {
			return tx.(*types.TbTransaction), nil
		}
		return nil, fmt.Errorf("transaction content for txID %v and hash %v does not exist", txID, hash)
	}
	return nil, fmt.Errorf("transaction with txID %v does not exist", txID)
}

// RemoveHashes deletes a series of transactions by their hash from TxStore. RemoveHashes can take a potentially large hash array, so it should be passed by reference.
func (t *TbTxStore) RemoveHashes(hashes *types.SHA256HashList, reEntryProtection ReEntryProtectionFlags, reason string) {
	for _, hash := range *hashes {
		t.remove(string(hash[:]), reEntryProtection, reason)
	}
}

// Iter returns a channel iterator for all transactions in TbTxStore
func (t *TbTxStore) Iter() <-chan *types.TbTransaction {
	txCh := make(chan *types.TbTransaction)
	go func() {
		for item := range t.hashToContent.IterBuffered() {
			tx := item.Val.(*types.TbTransaction)
			if t.clock.Now().Sub(tx.AddTime()) < t.maxTxAge {
				txCh <- tx
			}
		}
		close(txCh)
	}()
	return txCh
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

	// if txID was not provided, assign txID (if we are running as assigner) note that assigner.Next() provides TxIDEmpty if we are not assigning
	// also, txID is not assigned if transaction is validators_only if we assigned txID, result.AssignedTxID hold non TxIDEmpty value
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

type networkData struct {
	maxAge      time.Duration
	ages        []int
	cleanAge    int
	cleanNoTxID int
}

func (t *TbTxStore) clean() (cleaned int, cleanedTxIDs types.TxIDsByNetwork) {
	currTime := t.clock.Now()

	networks := make(map[types.NetworkNum]*networkData)
	cleanedTxIDs = make(types.TxIDsByNetwork)

	for item := range t.hashToContent.IterBuffered() {
		tbTransaction := item.Val.(*types.TbTransaction)
		netData, ok := networks[tbTransaction.NetworkNum()]
		if !ok {
			netData = &networkData{}
			networks[tbTransaction.NetworkNum()] = netData
		}
		txAge := int(currTime.Sub(tbTransaction.AddTime()) / time.Second)
		networks[tbTransaction.NetworkNum()].ages = append(networks[tbTransaction.NetworkNum()].ages, txAge)
	}

	for net, netData := range networks {
		// if we are below the number of allowed Txs, no need to do anything
		if len(netData.ages) <= TxStoreMaxSize {
			networks[net].maxAge = t.maxTxAge
			continue
		}
		// per network, sort ages in ascending order
		sort.Ints(netData.ages)
		// in order to avoid many cleanup messages, cleanup only 90% of the TxStoreMaxSize
		networks[net].maxAge = time.Duration(netData.ages[int(TxStoreMaxSize*0.9)-1]) * time.Second
		if networks[net].maxAge > t.maxTxAge {
			networks[net].maxAge = t.maxTxAge
		}
		log.Debugf("TxStore size for network %v is %v. Cleaning %v transactions older than %v",
			net, len(netData.ages), len(netData.ages)-TxStoreMaxSize, networks[net].maxAge)
	}

	for item := range t.hashToContent.IterBuffered() {
		tbTransaction := item.Val.(*types.TbTransaction)
		networkNum := tbTransaction.NetworkNum()
		removeReason := ""
		txAge := currTime.Sub(tbTransaction.AddTime())
		netData, ok := networks[networkNum]
		if ok && txAge > netData.maxAge {
			removeReason = fmt.Sprintf("transation age %v is greater than  %v", txAge, netData.maxAge)
			netData.cleanAge++
		} else {
			if txAge > t.noTxIDAge && len(tbTransaction.TxIDs()) == 0 {
				removeReason = fmt.Sprintf("transation age %v but no txID", txAge)
				netData.cleanNoTxID++
			}
		}
		if removeReason != "" {
			// remove the transaction by hash from both maps
			// no need to add the hash to the history as it is deleted after long time
			// add to hash history to prevent a lot of reentry (BSC, Polygon)
			t.remove(item.Key, FullReEntryProtection, removeReason)
			cleanedTxIDs[networkNum] = append(cleanedTxIDs[networkNum], tbTransaction.TxIDs()...)
		}
	}

	for net, netData := range networks {
		log.Debugf("TxStore network %v #txs before cleanup %v cleaned %v missing txID entries and %v aged entries",
			net, len(netData.ages), netData.cleanNoTxID, netData.cleanAge)
		cleaned += netData.cleanNoTxID + netData.cleanAge
	}

	return cleaned, cleanedTxIDs
}

// CleanNow performs an immediate cleanup of the TxStore
func (t *TbTxStore) CleanNow() {
	mapSizeBeforeClean := t.Count()
	startTime := t.clock.Now()
	cleaned, cleanedTxIDs := t.clean()
	log.Debugf("TxStore cleaned %v entries in %v. size before clean: %v size after clean: %v",
		cleaned, t.clock.Now().Sub(startTime), mapSizeBeforeClean, t.Count())
	if t.cleanedTxIDsChannel != nil && len(cleanedTxIDs) > 0 {
		t.cleanedTxIDsChannel <- cleanedTxIDs
	}
}

func (t *TbTxStore) cleanup() {
	ticker := t.clock.Ticker(t.cleanupFreq)
	for {
		select {
		case <-ticker.Alert():
			t.CleanNow()
		case <-t.quit:
			t.quit <- true
			ticker.Stop()
			return
		}
	}
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

// HasContent returns if a given transaction is in the transaction service
func (t *TbTxStore) HasContent(hash types.SHA256Hash) bool {
	tx, ok := t.Get(hash)
	if !ok {
		return false
	}
	return tx.Content() != nil
}

func (t *TbTxStore) refreshSeenTx(hash types.SHA256Hash) bool {
	if t.seenTxs.Exists(string(hash[:])) {
		t.seenTxs.Add(string(hash[:]), t.timeToAvoidReEntry)
		return true
	}
	return false
}
