package handler

import (
	"github.com/massbitprotocol/turbo/connections"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
)

// Relay represents a connection to relay node
type Relay struct {
	*TbConn
}

// NewOutboundRelay builds a new connection to a relay
func NewOutboundRelay(node connections.TbListener, sslCerts *utils.SSLCerts, relayIP string, relayPort int64, nodeId types.NodeID, connectionType utils.NodeType, clock utils.Clock) *Relay {
	return newRelay(node, func() (connections.Socket, error) {
		return connections.NewTLS(relayIP, int(relayPort), sslCerts)
	}, sslCerts, relayIP, relayPort, nodeId, connectionType, connections.LocalInitiatedPort, clock)
}

// NewInboundRelay builds a relay connection from a socket event initiated by a remote relay node
func NewInboundRelay(node connections.TbListener, socket connections.Socket, sslCerts *utils.SSLCerts, relayIP string, nodeID types.NodeID, relayType utils.NodeType, localPort int64, clock utils.Clock) *Relay {
	return newRelay(node,
		func() (connections.Socket, error) {
			return socket, nil
		},
		sslCerts, relayIP, connections.RemoteInitiatedPort, nodeID, relayType, localPort, clock)
}

func newRelay(node connections.TbListener, connect func() (connections.Socket, error), sslCerts *utils.SSLCerts, relayIP string, relayPort int64, nodeId types.NodeID, connectionType utils.NodeType, localPort int64, clock utils.Clock) *Relay {
	r := &Relay{}
	r.TbConn = NewTbConn(node, connect, r, sslCerts, relayIP, relayPort, nodeId, connectionType, true, localPort, clock)
	return r
}

// ProcessMessage handles messages received on the relay connection, delegating to the BxListener when appropriate
func (r *Relay) ProcessMessage(msg tbmessage.MessageBytes) {
	msgType := msg.TbType()
	if msgType != tbmessage.TxType {
		r.Log().Tracef("processing message %v", msgType)
	}

	switch msgType {
	case tbmessage.TxType:
		tx := &tbmessage.Tx{}
		_ = tx.Unpack(msg)
		_ = r.Node.HandleMsg(tx, r, connections.RunForeground)
	}
}
