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
	}, sslCerts, relayIP, relayPort, nodeId, connectionType, clock)
}

func newRelay(node connections.TbListener, connect func() (connections.Socket, error), sslCerts *utils.SSLCerts, relayIP string, relayPort int64, nodeId types.NodeID, connectionType utils.NodeType, clock utils.Clock) *Relay {
	r := &Relay{}
	r.TbConn = NewTbConn(node, connect, r, sslCerts, relayIP, relayPort, nodeId, connectionType, true, clock)
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
