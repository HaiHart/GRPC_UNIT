package tbmessage

import (
	"github.com/massbitprotocol/turbo/utils"
)

// ControlByteLen is the byte length of the control byte
const ControlByteLen = 1

// ValidControlByte is the final byte of all messages, indicating a fully packed message
const ValidControlByte = 0x01

// HeaderLen is the byte length of the common message headers
const HeaderLen = 20

// BroadcastHeaderLen is the byte length of the common broadcast message headers
const BroadcastHeaderLen = 72

// PayloadSizeOffset is the byte offset of the packed message size
const PayloadSizeOffset = 16

// TypeOffset is the byte offset of the packed message type
const TypeOffset = 4

// TypeLength is the byte length of the packed message type
const TypeLength = 12

// EncryptedTypeLen is the byte length of the encrypted byte
const EncryptedTypeLen = 1

// BroadcastTypeLen is the byte length of the broadcastType byte
const BroadcastTypeLen = 4

// Message type constants
const (
	TxType        = "tx"
	BroadcastType = "broadcast"
)

// SenderLen is the byte length of sender
const SenderLen = 20

// TimestampLen is the byte length of timestamps
const TimestampLen = 8

// ShortTimestampLen is the byte length of short timestamps
const ShortTimestampLen = 4

// SourceIDLen is the byte length of message source IDs
const SourceIDLen = 16

// NullByte is a character that is packed at the end of strings in buffers
const NullByte = "\x00"

var clock utils.Clock = utils.RealClock{}
