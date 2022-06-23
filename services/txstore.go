package services

import (
	"github.com/massbitprotocol/turbo/types"
	"time"
)

// TxStore is the service interface for transaction storage and processing
type TxStore interface {
	Add(hash types.SHA256Hash, content types.TxContent, txID types.TxID, networkNum types.NetworkNum, timestamp time.Time) TransactionResult
}

// TransactionResult is returned after the transaction service processes a new tx message, deciding whether to process it
type TransactionResult struct {
	NewTx            bool
	NewContent       bool
	NewTxID          bool
	Reprocess        bool
	FailedValidation bool
	Transaction      *types.TbTransaction
	AssignedTxID     types.TxID
	DebugData        interface{}
	AlreadySeen      bool
	Sender           types.Sender
}
