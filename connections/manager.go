package connections

// RelayInstruction specifies whether to connect or disconnect to the relay at an IP:Port
type RelayInstruction struct {
	IP   string
	Type ConnInstructionType
	Port int64
}

// ConnInstructionType specifies connection or disconnection
type ConnInstructionType int

const (
	// Connect is the instruction to connect to a relay
	Connect ConnInstructionType = iota
	// Disconnect is the instruction to disconnect from a relay
	Disconnect
)
