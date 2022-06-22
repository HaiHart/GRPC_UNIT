package tbmessage

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/massbitprotocol/turbo/types"
)

// Message is the base interface of all connection message sent on the wire
type Message interface {
	Pack() ([]byte, error)
	Unpack(buf []byte) error
	GetPriority() SendPriority
	SetPriority(priority SendPriority)
	String() string
}

// BroadcastMessage is the base interface of all broadcast message sent on the wire
type BroadcastMessage interface {
	Message
	GetNetworkNum() types.NetworkNum
}

// MessageBytes type def for message byte sets
type MessageBytes []byte

// TbType parses the message type from a bx message
func (mb MessageBytes) TbType() string {
	return string(bytes.Trim(mb[TypeOffset:TypeOffset+TypeLength], NullByte))
}

// String formats the message bytes for hex output
func (mb MessageBytes) String() string {
	return fmt.Sprintf("%v[%v]", mb.TbType(), hexutil.Encode(mb))
}
