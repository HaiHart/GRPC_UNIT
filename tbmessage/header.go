package tbmessage

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Header represents the shared header of a message
type Header struct {
	msgType string
}

// Pack serializes a Header into a buffer for sending on the wire
func (h *Header) Pack(buf *[]byte, msgType string) {
	h.msgType = msgType

	// TODO What does this "0xfffefdfc" value mean?
	binary.BigEndian.PutUint32(*buf, 0xfffefdfc)
	copy((*buf)[TypeOffset:], msgType)
	binary.LittleEndian.PutUint32((*buf)[PayloadSizeOffset:], uint32(len(*buf))-HeaderLen)

	(*buf)[len(*buf)-ControlByteLen] = ValidControlByte
}

// Unpack deserializes a Header from a buffer
func (h *Header) Unpack(buf []byte) error {
	h.msgType = string(bytes.Trim(buf[TypeOffset:TypeOffset+TypeLength], NullByte))
	return nil
}

// Size returns the byte length of header plus ControlByteLen
func (h *Header) Size() uint32 {
	return HeaderLen + ControlByteLen
}

func (h *Header) String() string {
	return fmt.Sprintf("Message<type: %v>", h.msgType)
}

func checkBufSize(buf *[]byte, offset int, size int) error {
	if len(*buf) < offset+size {
		return fmt.Errorf("invalid message format, %v bytes needed at offset %v but buff size is %v, buffer: %v",
			size, offset, len(*buf), hex.EncodeToString(*buf))
	}
	return nil
}
