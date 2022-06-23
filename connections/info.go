package connections

import (
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"time"
)

// Info represents various information fields about the connection.
type Info struct {
	NodeID          types.NodeID
	AccountID       types.AccountID
	PeerIP          string
	PeerPort        int64
	PeerEnode       string
	LocalPort       int64 // either the local listening server port, or 0 for outbound connections
	ConnectionType  utils.NodeType
	ConnectionState string
	NetworkNum      types.NetworkNum
	FromMe          bool
	Capabilities    types.CapabilityFlags
	Version         string
	ConnectedAt     time.Time
}

// IsGateway indicates if the connection is a gateway
func (ci Info) IsGateway() bool {
	return ci.ConnectionType&utils.Gateway != 0
}

// IsBDN indicates if the connection is a BDN gateway
func (ci Info) IsBDN() bool {
	return ci.Capabilities&types.CapabilityBDN != 0
}

// IsRelay indicates if the connection is a relay type
func (ci Info) IsRelay() bool {
	return ci.ConnectionType&utils.RelayTransaction != 0 || ci.ConnectionType&utils.RelayBlock != 0
}
