package tbmessage

import (
	"encoding/binary"
	"github.com/massbitprotocol/turbo/types"
)

// Broadcast represents the broadcast message
type Broadcast struct {
	BroadcastHeader
	broadcastType [BroadcastTypeLen]byte
	encrypted     bool
	block         []byte
	txIDs         types.TxIDList
}

// NewBlockBroadcast creates a new broadcast message containing block message bytes
func NewBlockBroadcast(hash types.SHA256Hash, block []byte, txIDs types.TxIDList, networkNum types.NetworkNum) *Broadcast {
	var broadcastType [BroadcastTypeLen]byte
	copy(broadcastType[:], "blok")
	b := &Broadcast{
		broadcastType: broadcastType,
		encrypted:     false,
		block:         block,
		txIDs:         txIDs,
	}
	b.SetHash(hash)
	b.SetNetworkNum(networkNum)
	return b
}

// BroadcastType returns the broadcast type
func (b Broadcast) BroadcastType() [BroadcastTypeLen]byte {
	return b.broadcastType
}

// SetBroadcastType sets the broadcast type
func (b *Broadcast) SetBroadcastType(broadcastType [BroadcastTypeLen]byte) {
	b.broadcastType = broadcastType
}

// Encrypted returns the encrypted byte
func (b Broadcast) Encrypted() bool {
	return b.encrypted
}

// SetEncrypted sets the encrypted byte
func (b *Broadcast) SetEncrypted(encrypted bool) {
	b.encrypted = encrypted
}

// Block returns the block
func (b Broadcast) Block() []byte {
	return b.block
}

// SetBlock sets the block
func (b *Broadcast) SetBlock(block []byte) {
	b.block = block
}

// TxIDs return block's transaction IDs
func (b Broadcast) TxIDs() types.TxIDList {
	return b.txIDs
}

// SetSids sets the sids
func (b *Broadcast) SetSids(txIDs types.TxIDList) {
	b.txIDs = txIDs
}

// Pack serializes a Broadcast into a buffer for sending
func (b *Broadcast) Pack() ([]byte, error) {
	bufLen := b.Size()
	buf := make([]byte, bufLen)
	b.BroadcastHeader.Pack(&buf, BroadcastType)
	offset := BroadcastHeaderLen
	copy(buf[offset:], b.broadcastType[:])
	offset += BroadcastTypeLen
	if b.encrypted {
		copy(buf[offset:], []uint8{1})
	} else {
		copy(buf[offset:], []uint8{0})
	}
	offset += EncryptedTypeLen
	binary.LittleEndian.PutUint64(buf[offset:], uint64(len(b.block)+types.UInt64Len))
	offset += types.UInt64Len
	copy(buf[offset:], b.block)
	offset += len(b.block)
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(b.txIDs)))
	offset += types.UInt32Len
	for _, sid := range b.txIDs {
		binary.LittleEndian.PutUint32(buf[offset:], uint32(sid))
		offset += types.UInt32Len
	}
	return buf, nil
}

// Unpack deserializes a Broadcast from a buffer
func (b *Broadcast) Unpack(buf []byte) error {
	if err := b.BroadcastHeader.Unpack(buf); err != nil {
		return err
	}

	offset := BroadcastHeaderLen
	copy(b.broadcastType[:], buf[offset:])
	offset += BroadcastTypeLen
	b.encrypted = int(buf[offset : offset+EncryptedTypeLen][0]) != 0
	offset += EncryptedTypeLen

	if err := checkBufSize(&buf, offset, types.UInt64Len); err != nil {
		return err
	}
	sidsOffset := int(binary.LittleEndian.Uint64(buf[offset:]))

	// sidsOffset includes its types.UInt64Len
	if err := checkBufSize(&buf, offset, sidsOffset); err != nil {
		return err
	}
	b.block = buf[offset+types.UInt64Len : offset+sidsOffset]
	offset += sidsOffset

	if err := checkBufSize(&buf, offset, types.UInt32Len); err != nil {
		return err
	}
	sidsLen := int(binary.LittleEndian.Uint32(buf[offset:]))
	offset += types.UInt32Len

	if err := checkBufSize(&buf, offset, types.UInt32Len*sidsLen); err != nil {
		return err
	}
	for i := 0; i < sidsLen; i++ {
		txID := types.TxID(binary.LittleEndian.Uint32(buf[offset:]))
		offset += types.UInt32Len
		b.txIDs = append(b.txIDs, txID)
	}

	return nil
}

// Size calculate msg size
func (b *Broadcast) Size() uint32 {
	return b.fixedSize() +
		types.UInt64Len + // sids offset
		uint32(len(b.block)) +
		types.UInt32Len + // sids len
		(uint32(len(b.txIDs)) * types.UInt32Len)
}

func (b *Broadcast) fixedSize() uint32 {
	return b.BroadcastHeader.Size() + BroadcastTypeLen + EncryptedTypeLen
}
