package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"fmt"
	"github.com/op/go-logging"
	"io"
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
	remoteAddr := conn.RemoteAddr()
	log.Info("[Net] New connection: %s", remoteAddr)

	for {
		closed, err := handleMessage(conn)
		if closed {
			break
		}
		if err != nil {
			log.Error("[Net] %s", err)
		}
	}

	log.Info("[Net] Connection closed: %s", remoteAddr)
}

func handleMessage(conn net.Conn) (bool, error) {
	// Read message length as uint32
	var msgLen uint32
	err := binary.Read(conn, binary.BigEndian, &msgLen)
	if err != nil {
		return err == io.EOF, fmt.Errorf("Could not read message length:", err)
	}

	log.Debug("[Message] Message length: %d bytes", msgLen)
	msgBytes := make([]byte, msgLen)
	_, err = conn.Read(msgBytes)
	if err != nil {
		return err == io.EOF, fmt.Errorf("Could not read message:", err)
	}

	msg := &cellaserv.Message{}
	err = proto.Unmarshal(msgBytes, msg)
	if err != nil {
		return false, fmt.Errorf("Could not unmarshal message:", err)
	}

	switch msg.GetType() {
	case cellaserv.Message_Register:
		return false, handleRegister(msg)
	default:
		return false, fmt.Errorf("Unknown message type: %d", msg.GetType())
	}
}

func handleRegister(msg *cellaserv.Message) error {
	return nil
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
