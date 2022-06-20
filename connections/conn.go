package connections

import (
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/tbmessage"
	"time"
)

// ConnHandler defines the methods needed to handle connections
type ConnHandler interface {
	ProcessMessage(msg tbmessage.MessageBytes)
}

// Conn defines a network interface that sends and receives messages
type Conn interface {
	ID() Socket
	Info() Info
	IsOpen() bool
	IsDisabled() bool

	Protocol() tbmessage.Protocol
	SetProtocol(tbmessage.Protocol)

	Log() *log.Entry

	Connect() error
	ReadMessages(callBack func(tbmessage.MessageBytes), readDeadline time.Duration, headerLen int, readPayloadLen func([]byte) int) (int, error)
	Send(msg tbmessage.Message) error
	SendWithDelay(msg tbmessage.Message, delay time.Duration) error
	Disable(reason string)
	Close(reason string) error
}

// ConnList represents the set of connections a node is maintaining
type ConnList []Conn

// MsgHandlingOptions represents background/foreground options for message handling
type MsgHandlingOptions bool

// MsgHandlingOptions enumeration
const (
	RunBackground MsgHandlingOptions = true
	RunForeground MsgHandlingOptions = false
)

// TbListener defines a struct that is capable of processing messages
type TbListener interface {
	HandleMsg(msg tbmessage.Message, conn Conn) error
	ValidateConnection(conn Conn) error

	// OnConnEstablished is a callback for when a connection has been connected and finished its handshake
	OnConnEstablished(conn Conn) error

	// OnConnClosed is a callback for when a connection is closed with no expectation of retrying
	OnConnClosed(conn Conn) error
}
