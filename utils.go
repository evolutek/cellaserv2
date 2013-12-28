package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"net"
)

func sendMessage(conn net.Conn, msg *cellaserv.Message) {
	log.Debug("[Net] Sending message to %s", conn.RemoteAddr())

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Error("[Net] Could not marshal outgoing message")
	}
	dumpOutgoing(msgBytes)

	msgLen := uint32(len(msgBytes))
	sendRawMessageLen(conn, msgLen, msgBytes)
}

func sendRawMessageLen(conn net.Conn, msgLen uint32, msg []byte) {
	// Any IO error will be detected by the main loop
	binary.Write(conn, binary.BigEndian, msgLen)
	conn.Write(msg)
}
