package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"time"
)

// TbBlockTransaction represents a tx in the TbBlock.
type TbBlockTransaction struct {
	hash    SHA256Hash
	content []byte
}

// NewTbBlockTransaction creates a new tx in the TbBlock. This transaction is usable for compression.
func NewTbBlockTransaction(hash SHA256Hash, content []byte) *TbBlockTransaction {
	return &TbBlockTransaction{
		hash:    hash,
		content: content,
	}
}

// NewRawTbBlockTransaction creates a new transaction that's not ready for compression. This should only be used when parsing the result of an existing TbBlock.
func NewRawTbBlockTransaction(content []byte) *TbBlockTransaction {
	return &TbBlockTransaction{
		content: content,
	}
}

// Hash returns the transaction hash
func (b TbBlockTransaction) Hash() SHA256Hash {
	return b.hash
}

// Content returns the transaction bytes
func (b TbBlockTransaction) Content() []byte {
	return b.content
}

// TbBlock represents an encoded block ready for compression or decompression
type TbBlock struct {
	hash            SHA256Hash
	Header          []byte
	Txs             []*TbBlockTransaction
	Trailer         []byte
	TotalDifficulty *big.Int
	Number          *big.Int
	timestamp       time.Time
	size            int
}

// NewTbBlock creates a new TbBlock that's ready for compression. This means that all transaction hashes must be included.
func NewTbBlock(hash SHA256Hash, header []byte, txs []*TbBlockTransaction, trailer []byte, totalDifficulty *big.Int, number *big.Int, size int) (*TbBlock, error) {
	for _, tx := range txs {
		if tx.Hash() == (SHA256Hash{}) {
			return nil, errors.New("all transactions must contain hashes")
		}
	}
	return NewRawTbBlock(hash, header, txs, trailer, totalDifficulty, number, size), nil
}

// NewRawTbBlock create a new TbBlock without compression restrictions. This should only be used when parsing the result of an existing TbBlock.
func NewRawTbBlock(hash SHA256Hash, header []byte, txs []*TbBlockTransaction, trailer []byte, totalDifficulty *big.Int, number *big.Int, size int) *TbBlock {
	TbBlock := &TbBlock{
		hash:            hash,
		Header:          header,
		Txs:             txs,
		Trailer:         trailer,
		TotalDifficulty: totalDifficulty,
		Number:          number,
		timestamp:       time.Now(),
		size:            size,
	}
	return TbBlock
}

// Serialize returns an expensive string representation of the TbBlock
func (b TbBlock) Serialize() string {
	m := make(map[string]interface{})
	m["header"] = hex.EncodeToString(b.Header)
	m["trailer"] = hex.EncodeToString(b.Trailer)
	m["totalDifficulty"] = b.TotalDifficulty.String()
	m["number"] = b.Number.String()

	txs := make([]string, 0, len(b.Txs))
	for _, tx := range b.Txs {
		txs = append(txs, hex.EncodeToString(tx.content))
	}
	m["txs"] = txs

	jsonBytes, _ := json.Marshal(m)
	return string(jsonBytes)
}

// Hash returns block hash
func (b TbBlock) Hash() SHA256Hash {
	return b.hash
}

// Timestamp returns block add time
func (b TbBlock) Timestamp() time.Time {
	return b.timestamp
}

// Size returns the original blockchain block
func (b TbBlock) Size() int {
	return b.size
}

// SetSize sets the original blockchain block
func (b *TbBlock) SetSize(size int) {
	b.size = size
}

// Equals checks the byte contents of each part of the provided TbBlock. Note that some fields are set throughout the object's lifecycle (block hash, transaction hash), so these fields are not checked for equality.
func (b *TbBlock) Equals(other *TbBlock) bool {
	if !bytes.Equal(b.Header, other.Header) || !bytes.Equal(b.Trailer, other.Trailer) {
		return false
	}

	for i, tx := range b.Txs {
		otherTx := other.Txs[i]
		if !bytes.Equal(tx.content, otherTx.content) {
			return false
		}
	}

	return b.TotalDifficulty.Cmp(other.TotalDifficulty) == 0 && b.Number.Cmp(other.Number) == 0
}
