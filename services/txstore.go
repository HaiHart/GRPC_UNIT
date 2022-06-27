package services

import (
	"github.com/massbitprotocol/turbo/types"
	"time"
)

// ReEntryProtectionFlags protect against hash re-entrance
type ReEntryProtectionFlags uint8

// flag constant values
const (
	NoReEntryProtection ReEntryProtectionFlags = iota
	ShortReEntryProtection
	FullReEntryProtection
)

// ShortReEntryProtectionDuration defines the short duration for TxStore reentry protection
const ShortReEntryProtectionDuration = 30 * time.Minute

// TxStore is the service interface for transaction storage and processing
type TxStore interface {
	Start() error
	Stop()

	Add(hash types.SHA256Hash, content types.TxContent, txID types.TxID, networkNum types.NetworkNum, timestamp time.Time) TransactionResult
	Get(hash types.SHA256Hash) (*types.TbTransaction, bool)
	Known(hash types.SHA256Hash) bool
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
