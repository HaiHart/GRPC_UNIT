package types

import (
	"sync"
	"time"
)

// TxContent represents a byte array containing full transaction bytes
type TxContent []byte

// TbTransaction represents a Turbo transaction
type TbTransaction struct {
	m          sync.Mutex
	hash       SHA256Hash
	content    TxContent
	shortIDs   ShortIDList
	addTime    time.Time
	flags      TxFlags
	networkNum NetworkNum
}
