package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/massbitprotocol/turbo/tbmessage"
	"log"
	"net"
)

func main() {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair("datadir/gateway/private/gateway_cert.pem", "datadir/gateway/private/gateway_key.pem")
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := bytes.Buffer{}
	packet := make([]byte, 4096)
	for {
		n, err := conn.Read(packet)
		if err != nil {
		}
		buf.Write(packet[:n])
		for {
			bufLen := buf.Len()
			if bufLen < tbmessage.HeaderLen {
				break
			}

			payloadLen := int(binary.LittleEndian.Uint32(buf.Bytes()[tbmessage.PayloadSizeOffset:]))
			if bufLen < tbmessage.HeaderLen+payloadLen {
				break
			}

			// allocate an array for the message to protect from overrun by multiple go routines
			msg := make([]byte, tbmessage.HeaderLen+payloadLen)
			_, err = buf.Read(msg)
			if err != nil {

			}
			processMsg(msg)
		}
	}
}

func processMsg(msg tbmessage.MessageBytes) {
	msgType := msg.TbType()
	if msgType != tbmessage.TxType {
	}

	switch msgType {
	case tbmessage.TxType:
		tx := &tbmessage.Tx{}
		_ = tx.Unpack(msg)
		fmt.Printf("received transaction %v\n", tx.Hash())

	case tbmessage.BroadcastType:
		block := &tbmessage.Broadcast{}
		_ = block.Unpack(msg)
		fmt.Printf("received block %v\n", block.Hash())
	}
}
