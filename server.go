package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"fmt"
	"github.com/op/go-logging"
	"net"
)

// Setup log
var log = logging.MustGetLogger("cellaserv")

type Service struct {
	ip net.IPAddr
}

// Currently connected services
var services []Service

// Each connection is managed here
func handle(conn net.Conn) {
	for {
		handleMessage(conn)
	}
}

func handleMessage(conn net.Conn) error {
	remoteAddr := conn.RemoteAddr()
	log.Debug("New connection from %s", remoteAddr)

	// Read message length as uint32
	var msgLen uint32
	err := binary.Read(conn, binary.BigEndian, &msgLen)
	if err != nil {
		return fmt.Errorf("Could not read message length:", err)
	}

	log.Debug("Message lenght: %d bytes", msgLen)
	msgBytes := make([]byte, msgLen)
	_, err = conn.Read(msgBytes)
	if err != nil {
		return fmt.Errorf("Could not read message:", err)
	}

	msg := &cellaserv.Message{}
	err = proto.Unmarshal(msgBytes, msg)
	if err != nil {
		return fmt.Errorf("Could not unmarshal message:", err)
	}

	switch msg.GetType() {
	case cellaserv.Message_Register:
		handleRegister(msg)
	default:
		return fmt.Errorf("Unknown message type: %d", msg.GetType())
	}

	return nil
}

func handleRegister(msg *cellaserv.Message) {
}

func serve() {
	ln, err := net.Listen("tcp", ":4200")
	if err != nil {
		log.Error("error")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go handle(conn)
	}
}

func main() {
	serve()
	// Will add internal "cellaserv" service setup here
}
