package types

import (
	"encoding/hex"
	"fmt"
)

// UInt32Len is the byte length of unsigned 32bit integers
const UInt32Len = 4

// NodeEndpoint - represent the node endpoint struct sent in BdnPerformanceStats
type NodeEndpoint struct {
	IP        string
	Port      int
	PublicKey string
}

// String returns string representation of NodeEndpoint
func (e NodeEndpoint) String() string {
	return fmt.Sprintf("%v %v %v", e.IP, e.Port, e.PublicKey)
}

// IPPort returns string of IP and Port
func (e NodeEndpoint) IPPort() string {
	return fmt.Sprintf("%v %v", e.IP, e.Port)
}

// TxID represents the transaction ID
type TxID uint32

// TxIDList represents transaction ID list
type TxIDList []TxID

// TxIDsByNetwork represents map of txIDs by network
type TxIDsByNetwork map[NetworkNum]TxIDList

// NodeID represents a node's assigned ID. This field is a UUID.
type NodeID string

// Sender represents sender type
type Sender [20]byte

// String returns string of the Sender
func (s Sender) String() string {
	return hex.EncodeToString(s[:])
}

// TxIDEmpty is the default value indicating no assigned short ID
const TxIDEmpty = 0

// TxIDLen is the byte length of packed short IDs
const TxIDLen = UInt32Len

// NetworkNum represents the network that a message is being routed in (Ethereum Mainnet, Ethereum Ropsten, etc.)
type NetworkNum uint32

// NetworkNumLen is the byte length of packed network numbers
const NetworkNumLen = UInt32Len

// AllNetworkNum is the network number for relays that facilitate transactions from all networks
const AllNetworkNum NetworkNum = 0
