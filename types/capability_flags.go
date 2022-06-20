package types

// CapabilityFlags represents various flags for capabilities in hello msg
type CapabilityFlags uint16

const (
	CapabilityBDN CapabilityFlags = 1 << iota
)
