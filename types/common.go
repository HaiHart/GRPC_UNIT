package types

import "fmt"

// UInt32Len is the byte length of unsigned 32bit integers
const UInt32Len = 4

// UInt64Len is the byte length of unsigned 64bit integers
const UInt64Len = 8

// UInt16Len is the byte length of unsigned 16bit integers
const UInt16Len = 2

// UInt8Len is the byte length of unsigned 8bit integers
const UInt8Len = 1

// TxFlagsLen represents the byte length of transaction flag
const TxFlagsLen = 2

// NodeEndpoint represent node endpoint struct sent in network
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

// ShortID represents the compressed transaction ID
type ShortID uint32

// ShortIDList represents short id list
type ShortIDList []ShortID

// ShortIDsByNetwork represents map of shortIDs by network
type ShortIDsByNetwork map[NetworkNum]ShortIDList

// NetworkNum represents the network that a message is being routed in (Ethereum Mainnet, Ethereum Ropsten, etc.)
type NetworkNum uint32

// NetworkNumLen is the byte length of packed network numbers
const NetworkNumLen = UInt32Len
