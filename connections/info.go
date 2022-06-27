package connections

import (
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"time"
)

// Info represents various information fields about the connection.
type Info struct {
	NodeID         types.NodeID
	PeerIP         string
	PeerPort       int64
	PeerEnode      string
	LocalPort      int64 // either the local listening server port, or 0 for outbound connections
	ConnectionType utils.NodeType
	NetworkNum     types.NetworkNum
	FromMe         bool
	Version        string
	ConnectedAt    time.Time
}

// IsGateway indicates if the connection is a gateway
func (ci Info) IsGateway() bool {
	return ci.ConnectionType&utils.Gateway != 0
}

// IsRelay indicates if the connection is a relay type
func (ci Info) IsRelay() bool {
	return ci.ConnectionType&utils.RelayTransaction != 0 || ci.ConnectionType&utils.RelayBlock != 0
}
