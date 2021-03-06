package nodes

import (
	"fmt"
	"github.com/massbitprotocol/turbo/config"
	"github.com/massbitprotocol/turbo/connections"
	"github.com/massbitprotocol/turbo/tbmessage"
	"github.com/massbitprotocol/turbo/types"
	"github.com/massbitprotocol/turbo/utils"
	"sync"
)

// Base is a base handler for nodes
type Base struct {
	Abstract
	Config          *config.TurboConfig
	ConnectionsLock *sync.RWMutex
	Connections     connections.ConnList
	dataDir         string
	clock           utils.RealClock
}

// NewBase initializes a generic Base
func NewBase(config *config.TurboConfig, dataDir string) Base {
	return Base{
		Config:          config,
		Connections:     make(connections.ConnList, 0),
		ConnectionsLock: &sync.RWMutex{},
		dataDir:         dataDir,
		clock:           utils.RealClock{},
	}
}

// OnConnEstablished - a callback function called when new connection is established
func (b *Base) OnConnEstablished(conn connections.Conn) error {
	connInfo := conn.Info()
	conn.Log().Infof("connection established, gateway: %v, protocol version %v, network %v, on local port %v",
		connInfo.IsGateway(), connInfo.NetworkNum, conn.Info().LocalPort)
	b.ConnectionsLock.Lock()
	defer b.ConnectionsLock.Unlock()
	b.Connections = append(b.Connections, conn)
	return nil
}

// ValidateConnection - validates connection
func (b *Base) ValidateConnection(_ connections.Conn) error {
	return nil
}

// OnConnClosed - a callback function. Called when new connection is closed
func (b *Base) OnConnClosed(conn connections.Conn) error {
	b.ConnectionsLock.Lock()
	defer b.ConnectionsLock.Unlock()
	for idx, connection := range b.Connections {
		if connection.ID() == conn.ID() {
			b.Connections = append(b.Connections[:idx], b.Connections[idx+1:]...)
			conn.Log().Debugf("connection closed and removed from connection pool")
			return nil
		}
	}
	err := fmt.Errorf("connection can't be removed from connection list - not found")
	conn.Log().Debug(err)
	return err
}

// HandleMsg - a callback function. Generic handling for common messages
func (b *Base) HandleMsg(msg tbmessage.Message, source connections.Conn) error {
	switch msg.(type) {

	}
	return nil
}

// DisconnectConn - disconnect a specific connection
func (b *Base) DisconnectConn(id types.NodeID) {
	b.ConnectionsLock.Lock()
	defer b.ConnectionsLock.Unlock()
	for _, conn := range b.Connections {
		if id == conn.Info().NodeID {
			// closing in a new go routine in order to avoid deadlock while Close method acquiring ConnectionsLock
			go func() {
				err := conn.Close("disconnect requested")
				if err != nil {
					conn.Log().Debug(err)
				}
			}()
		}
	}
}
