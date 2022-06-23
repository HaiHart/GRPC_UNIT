package tbmessage

import (
	"encoding/binary"
	"github.com/massbitprotocol/turbo/types"
	"time"
)

// Tx is the message struct that carries a transaction from a specific blockchain
type Tx struct {
	BroadcastHeader
	txID      types.TxID
	timestamp time.Time
	content   []byte
	sender    types.Sender
}

// NewTx constructs a new transaction message, using the provided hash function on the transaction contents to determine the hash
func NewTx(hash types.SHA256Hash, content []byte, networkNum types.NetworkNum) *Tx {
	tx := &Tx{}
	tx.SetNetworkNum(networkNum)
	tx.SetContent(content)
	tx.SetHash(hash)
	tx.SetTimestamp(clock.Now())
	return tx
}

// HashString returns the hex encoded Hash, optionally with 0x prefix
func (m *Tx) HashString(prefix bool) string {
	return m.hash.Format(prefix)
}

// TxID returns the assigned transaction ID
func (m *Tx) TxID() types.TxID {
	return m.txID
}

// SetTxID sets the assigned transaction ID
func (m *Tx) SetTxID(txID types.TxID) {
	m.txID = txID
}

// Content returns the blockchain transaction bytes
func (m *Tx) Content() (content []byte) {
	return m.content
}

// SetContent sets the transaction content
func (m *Tx) SetContent(content []byte) {
	m.content = content
}

// Timestamp indicates when the TxMessage was sent.
func (m *Tx) Timestamp() time.Time {
	return m.timestamp
}

// SetTimestamp sets the message creation timestamp
func (m *Tx) SetTimestamp(timestamp time.Time) {
	m.timestamp = timestamp
}

// Sender return tx sender
func (m *Tx) Sender() types.Sender {
	return m.sender
}

// SetSender sets the sender
func (m *Tx) SetSender(sender types.Sender) {
	copy(m.sender[:], sender[:])
}

// Pack serializes a Tx into a buffer for sending
func (m Tx) Pack() ([]byte, error) {
	bufLen := m.size()
	buf := make([]byte, bufLen)
	m.BroadcastHeader.Pack(&buf, TxType)

	offset := BroadcastHeaderLen

	binary.LittleEndian.PutUint32(buf[offset:], uint32(m.txID))
	offset += types.TxIDLen

	timestamp := m.timestamp.UnixNano() >> 10
	binary.LittleEndian.PutUint32(buf[offset:], uint32(timestamp))
	offset += ShortTimestampLen

	// should include sender and nonce only if content is included
	if len(m.content) > 0 {
		copy(buf[offset:], m.content)
		offset += len(m.content)
		copy(buf[offset:], m.sender[:])
	}
	return buf, nil
}

// Unpack deserializes a Tx from a buffer
func (m *Tx) Unpack(buf []byte) error {
	if err := m.BroadcastHeader.Unpack(buf); err != nil {
		return err
	}

	offset := BroadcastHeaderLen
	m.txID = types.TxID(binary.LittleEndian.Uint32(buf[offset:]))
	offset += types.TxIDLen

	timestamp := binary.LittleEndian.Uint32(buf[offset:])
	offset += ShortTimestampLen
	toMerge := uint64(timestamp) << 10
	// the leading 22 bits were wiped out during pack(), the trailing 10 bits (nanoseconds) were dropped out as well
	// 0xFFFFFC0000000000 is hex for 0b1111111111111111111111000000000000000000000000000000000000000000
	// leading 22 1s, following by 42 0s
	now := clock.Now().UnixNano()
	nanoseconds := toMerge | uint64(now)&0xFFFFFC0000000000
	result := int64(nanoseconds)
	// There are two edge cases we need to check
	// 1. overflow, the Tx message only track 32 bit, if the sender time is 1234+FFFF..FF(32bit)+(10bits offset, not important)
	// the receiver time is 1235+00000(32bit)+(10 bits offset,not important), after the merge we will end up with 1235+FFF..FFF+..
	// the overflow will result in roughly 1 hour longer than expected.
	// To solve this, we need to compare the current time with the unpacked timestamp, if the gap is huge ( more than an hour )
	// then overflow happened, and we need to subtract the timestamp by 0x40000000000 nanoseconds
	if result-now >= time.Hour.Nanoseconds() {
		// overflow
		result = result - 0x40000000000
	}
	// 2. underflow, 1234+0000..000+10bits for the sender, and 1233+FFFFF..FF+10bits due to the clock time difference of the machine.
	// after the merge this can become 1233+000000..00+10bit which is far older than expected, so we need to add 0x40000000000 to the
	// timestamp
	if now-result >= time.Hour.Nanoseconds() {
		// underflow
		result = result + 0x40000000000
	}
	m.timestamp = time.Unix(0, result)

	m.content = buf[offset : len(buf)-SenderLen-ControlByteLen]
	offset += len(m.content)
	copy(m.sender[:], buf[offset:])

	return nil
}

func (m *Tx) size() uint32 {
	contentSize := uint32(len(m.content))
	if contentSize > 0 {
		contentSize += SenderLen
	}
	return m.BroadcastHeader.Size() + contentSize
}
