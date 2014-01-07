package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"net"
)

func sendReply(conn net.Conn, req *cellaserv.Request, data []byte) {
	rep := &cellaserv.Reply{Id: req.Id, Data: data}
	repBytes, err := proto.Marshal(rep)
	if err != nil {
		log.Error("[Message] Could not marshal outgoing reply")
	}

	msgType := cellaserv.Message_Reply
	msg := &cellaserv.Message{Type: &msgType, Content: repBytes}

	sendMessage(conn, msg)
}

func sendReplyError(conn net.Conn, req *cellaserv.Request, err_t cellaserv.Reply_Error_Type) {
	err := &cellaserv.Reply_Error{Type: &err_t}

	reply := &cellaserv.Reply{Error: err, Id: req.Id}
	replyBytes, _ := proto.Marshal(reply)

	msgType := cellaserv.Message_Reply
	msg := &cellaserv.Message{
		Type:    &msgType,
		Content: replyBytes,
	}
	sendMessage(conn, msg)
}

func sendMessage(conn net.Conn, msg *cellaserv.Message) {
	log.Debug("[Net] Sending message to %s", conn.RemoteAddr())

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Error("[Message] Could not marshal outgoing message")
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

// vim: set nowrap tw=100 noet sw=8:
