package handler

import (
	"encoding/binary"
	"github.com/massbitprotocol/turbo/connections"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"sync"
	"time"
)

const (
	connTimeout = 5 * time.Second
)

// TbConn is a connection to any other node. TbConn implements connections.ConnHandler.
type TbConn struct {
	connections.Conn

	Node    connections.TbListener
	Handler connections.ConnHandler

	lock                  *sync.Mutex
	connectionEstablished bool
	closed                bool
	nodeID                types.NodeID
	peerID                types.NodeID
	connectionType        utils.NodeType
	localPort             int64
	log                   *log.Entry
	clock                 utils.Clock
	connectedAt           time.Time
}

// NewTbConn constructs a connection to a turbo node
func NewTbConn(node connections.TbListener, connect func() (connections.Socket, error), handler connections.ConnHandler, sslCerts *utils.SSLCerts, ip string, port int64, nodeID types.NodeID, connectionType utils.NodeType, logMessages bool, localPort int64, clock utils.Clock) *TbConn {
	tc := &TbConn{
		Conn:      connections.NewSSLConnection(connect, sslCerts, ip, port, logMessages, 100000, clock),
		Node:      node,
		Handler:   handler,
		nodeID:    nodeID,
		localPort: localPort,
		log: log.WithFields(log.Fields{
			"connType":   connectionType.String(),
			"remoteAddr": "<connecting>",
		}),
		clock: clock,
	}
	// TODO: remove this
	tc.setConnectionEstablished()
	return tc
}

// Start kicks off main goroutine of the connection
func (b *TbConn) Start() error {
	go b.readLoop()
	return nil
}

// ProcessMessage constructs a message from the buffer and handles it
// This method only handles messages that do not require querying the BxListener interface
func (b *TbConn) ProcessMessage(msg tbmessage.MessageBytes) {
	msgType := msg.TbType()
	switch msgType {
	default:
		b.Log().Debugf("read %v (%d bytes)", msgType, len(msg))
	}
}

func (b *TbConn) setConnectionEstablished() {
	b.connectionEstablished = true
	_ = b.Node.OnConnEstablished(b)
}

// readLoop connects and reads messages from the socket.
// If we are the initiator of the connection we auto-recover on disconnect.
func (b *TbConn) readLoop() {
	isInitiator := b.Info().FromMe
	for {
		err := b.Connect()
		if err != nil {
			b.Log().Errorf("encountered connection error while connecting: %v", err)
			reason := "could not connect to remote"
			if !isInitiator {
				_ = b.Close(reason)
				break
			}

			_ = b.closeWithRetry(reason)
			// sleep before next connection attempt
			b.clock.Sleep(connTimeout)
			continue
		}

		if isInitiator {
		}

		closeReason := "read loop closed"
		for b.Conn.IsOpen() {
			// TODO remove this
			time.Sleep(10 * time.Minute)
			_, err = b.ReadMessages(b.Handler.ProcessMessage, 30*time.Second, tbmessage.HeaderLen, func(b []byte) int {
				return int(binary.LittleEndian.Uint32(b[tbmessage.PayloadSizeOffset:]))
			})
			if err != nil {
				closeReason = err.Error()
				b.Log().Tracef("connection closed: %v", err)
				break
			}
		}
		if !isInitiator {
			_ = b.Close(closeReason)
			break
		}

		_ = b.closeWithRetry(closeReason)
		if b.closed {
			break
		}

		// sleep before next connection attempt
		// note - in docker environment the docker-proxy may keep the port open after the docker was stopped. we
		// need this sleep to avoid fast connect/disconnect loop
		b.clock.Sleep(connTimeout)
	}
}

// closeWithRetry does not shut down the main go routines present in TbConn, only the ones in the ssl connection, which can be restarted on the next Connect
func (b *TbConn) closeWithRetry(reason string) error {
	b.connectionEstablished = false
	_ = b.Node.OnConnClosed(b)
	return b.Conn.Close(reason)
}
