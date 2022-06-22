package services

import (
	"github.com/massbitprotocol/turbo/connections"
	log "github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/tbmessage"
)

// AsyncMsgChannelSize - size of async message channel
const AsyncMsgChannelSize = 500

// MsgInfo wraps a turbo message and its source connection
type MsgInfo struct {
	Msg    tbmessage.Message
	Source connections.Conn
}

// AsyncMsgHandler handles messages asynchronously
type AsyncMsgHandler struct {
	AsyncMsgChan chan MsgInfo
	listener     connections.TbListener
}

func NewAsyncMsgChannel(listener connections.TbListener) chan MsgInfo {
	handler := &AsyncMsgHandler{
		AsyncMsgChan: make(chan MsgInfo, AsyncMsgChannelSize),
		listener:     listener,
	}
	go handler.HandleMsgAsync()
	return handler.AsyncMsgChan
}

// HandleMsgAsync handles messages pushed into the channel of AsyncMsgHandler
func (handler AsyncMsgHandler) HandleMsgAsync() {
	for {
		messageInfo, ok := <-handler.AsyncMsgChan
		if !ok {
			log.Error("unexpected termination of AsyncMsgHandler. AsyncMsgChannel was closed.")
			return
		}
		log.Tracef("async handling of %v from %v", messageInfo.Msg, messageInfo.Source)
		_ = handler.listener.HandleMsg(messageInfo.Msg, messageInfo.Source, connections.RunForeground)
	}
}
