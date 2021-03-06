package connections

import (
	"fmt"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"time"
)

// Blockchain is a placeholder struct to represent a connection for blockchain nodes
type Blockchain struct {
	endpoint types.NodeEndpoint
	log      *log.Entry
}

var _ Conn = (*Blockchain)(nil)

var blockchainTLSPlaceholder = TLS{}

// NewBlockchainConn return a new instance of the Blockchain placeholder connection
func NewBlockchainConn(ipEndpoint types.NodeEndpoint) Blockchain {
	return Blockchain{
		endpoint: ipEndpoint,
		log: log.WithFields(log.Fields{
			"connType":   utils.Blockchain.String(),
			"remoteAddr": fmt.Sprintf("%v:%v", ipEndpoint.IP, ipEndpoint.Port),
		}),
	}
}

// NodeEndpoint return the blockchain connection endpoint
func (b Blockchain) NodeEndpoint() types.NodeEndpoint {
	return b.endpoint
}

// ID returns placeholder
func (b Blockchain) ID() Socket {
	return blockchainTLSPlaceholder
}

// Info returns connection metadata
func (b Blockchain) Info() Info {
	return Info{
		ConnectionType: utils.Blockchain,
		NetworkNum:     types.AllNetworkNum,
		PeerIP:         b.endpoint.IP,
		PeerPort:       int64(b.endpoint.Port),
		PeerEnode:      b.endpoint.PublicKey,
	}
}

// IsOpen is never true, since the Blockchain is not writable
func (b Blockchain) IsOpen() bool {
	return false
}

// IsDisabled indicates that Blockchain is never disabled
func (b Blockchain) IsDisabled() bool {
	return false
}

// Log returns the blockchain connection logger
func (b Blockchain) Log() *log.Entry {
	return b.log
}

// Connect is a no-op
func (b Blockchain) Connect() error {
	return nil
}

// ReadMessages is a no-op
func (b Blockchain) ReadMessages(callBack func(tbmessage.MessageBytes), readDeadline time.Duration, headerLen int, readPayloadLen func([]byte) int) (int, error) {
	return 0, nil
}

// Send is a no-op
func (b Blockchain) Send(msg tbmessage.Message) error {
	return nil
}

// SendWithDelay is a no-op
func (b Blockchain) SendWithDelay(msg tbmessage.Message, delay time.Duration) error {
	return nil
}

// Close is a no-op
func (b Blockchain) Close(reason string) error {
	return nil
}

// Disable is a no-op
func (b Blockchain) Disable(reason string) {
	return
}

// String returns the formatted representation of this placeholder connection
func (b Blockchain) String() string {
	return fmt.Sprintf("Blockchain %v", b.endpoint.String())
}
