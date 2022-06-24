package services

import (
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"math/big"
	"sync"
	"time"
)

// error constants for identifying special processing cases
var (
	ErrAlreadyProcessed = errors.New("already processed")
	ErrMissingShortIDs  = errors.New("missing SIDs")
)

// BxBlockConverter is the service interface for converting broadcast messages to/from bx blocks
type BxBlockConverter interface {
	BxBlockToBroadcast(*types.TbBlock, types.NetworkNum) (*tbmessage.Broadcast, types.TxIDList, error)
}

// BlockProcessor is the service interface for processing broadcast messages
type BlockProcessor interface {
	BxBlockConverter
}

type bxCompressedTransaction struct {
	IsFullTransaction bool
	Transaction       []byte
}

type bxBlockRLP struct {
	Header          rlp.RawValue
	Txs             []bxCompressedTransaction
	Trailer         rlp.RawValue
	TotalDifficulty *big.Int
	Number          *big.Int
}

type rlpBlockProcessor struct {
	txStore         TxStore
	processedBlocks HashHistory
	lock            *sync.Mutex
}

// NewRLPBlockProcessor returns a BlockProcessor for Ethereum blocks encoded in broadcast messages
func NewRLPBlockProcessor(txStore TxStore) BlockProcessor {
	bp := &rlpBlockProcessor{
		txStore:         txStore,
		processedBlocks: NewHashHistory("processedBlocks", 30*time.Minute),
		lock:            &sync.Mutex{},
	}
	return bp
}

func (bp *rlpBlockProcessor) BxBlockToBroadcast(block *types.TbBlock, networkNum types.NetworkNum) (*tbmessage.Broadcast, types.TxIDList, error) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	blockHash := block.Hash()
	if !bp.ShouldProcess(blockHash) {
		return nil, nil, ErrAlreadyProcessed
	}

	usedShortIDs := make(types.TxIDList, 0)
	txs := make([]bxCompressedTransaction, 0, len(block.Txs))
	// compress transactions in block if short ID is known
	for _, tx := range block.Txs {
		txs = append(txs, bxCompressedTransaction{
			IsFullTransaction: true,
			Transaction:       tx.Content(),
		})
	}

	rlpBlock := bxBlockRLP{
		Header:          block.Header,
		Txs:             txs,
		Trailer:         block.Trailer,
		TotalDifficulty: block.TotalDifficulty,
		Number:          block.Number,
	}
	encodedBlock, err := rlp.EncodeToBytes(rlpBlock)
	if err != nil {
		return nil, usedShortIDs, err
	}

	bp.markProcessed(blockHash)
	broadcastMessage := tbmessage.NewBlockBroadcast(block.Hash(), encodedBlock, usedShortIDs, networkNum)
	return broadcastMessage, usedShortIDs, nil
}

func (bp *rlpBlockProcessor) ShouldProcess(hash types.SHA256Hash) bool {
	return !bp.processedBlocks.Exists(hash.String())
}

func (bp *rlpBlockProcessor) markProcessed(hash types.SHA256Hash) {
	bp.processedBlocks.Add(hash.String(), 10*time.Minute)
}
