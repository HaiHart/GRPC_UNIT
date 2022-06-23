package services

import "github.com/massbitprotocol/turbo/types"

// TxIDAssigner - an interface for short ID assigner struct
type TxIDAssigner interface {
	Next() types.TxID
}

type emptyTxIDAssigner struct {
}

func (empty *emptyTxIDAssigner) Next() types.TxID {
	return types.TxIDEmpty
}

// NewEmptyTxIDAssigner - create an assigner that never assign
func NewEmptyTxIDAssigner() TxIDAssigner {
	return &emptyTxIDAssigner{}
}
