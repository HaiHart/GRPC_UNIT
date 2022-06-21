package tbmessage

import (
	"github.com/massbitprotocol/turbo/types"
)

// Tx is the message struct that carries a transaction from a specific blockchain
type Tx struct {
	BroadcastHeader
	content []byte
	sender  types.Sender
}

// NewTx constructs a new transaction message, using the provided hash function on the transaction contents to determine the hash
func NewTx(hash types.SHA256Hash, content []byte, networkNum types.NetworkNum) *Tx {
	tx := &Tx{}

	tx.SetNetworkNum(networkNum)
	tx.SetContent(content)
	tx.SetHash(hash)
	return tx
}

// SetContent sets the transaction content
func (m *Tx) SetContent(content []byte) {
	m.content = content
}

// Pack serializes a Tx into a buffer for sending
func (m Tx) Pack(protocol Protocol) ([]byte, error) {
	bufLen := m.size(protocol)
	buf := make([]byte, bufLen)
	m.BroadcastHeader.Pack(&buf, TxType)
	offset := BroadcastHeaderLen

	// should include sender and nonce only if content is included
	if len(m.content) > 0 {
		copy(buf[offset:], m.content)
		offset += len(m.content)
		copy(buf[offset:], m.sender[:])
	}

	return buf, nil
}

func (m *Tx) size(protocol Protocol) uint32 {
	contentSize := uint32(len(m.content))
	if contentSize > 0 {
		contentSize += SenderLen
	}
	return m.BroadcastHeader.Size() + contentSize
}
