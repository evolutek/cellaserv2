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

// Currently connected services
var services map[string]map[string]*Service
var servicesConn map[net.Conn][]*Service

// Manage incoming connexions
func handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	log.Info("[Net] New connection: %s", remoteAddr)

	// Handle all messages received on this connection
	for {
		closed, err := handleMessage(conn)
		if closed {
			break
		}
		if err != nil {
			log.Error("[Net] %s", err)
		}
	}

	// Remove services registered by this connection
	// TODO: notify goroutines waiting for acks for this service
	for _, s := range servicesConn[conn] {
		log.Info("[Services] Remove %+v", s)
		delete(services[s.name], s.identification)
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
		register := &cellaserv.Register{}
		err = proto.Unmarshal(msg.GetContent(), register)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal register:", err)
		}
		return false, handleRegister(conn, register)
	default:
		return false, fmt.Errorf("Unknown message type: %d", msg.GetType())
	}
}

// Start listening and receiving connections
func serve() {
	ln, err := net.Listen("tcp", ":4200")
	if err != nil {
		log.Error("[Net] Could not listen: %s", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("[Net] Could not accept: %s", err)
			continue
		}

		go handle(conn)
	}
}

func main() {
	// Initialize our maps
	services = make(map[string]map[string]*Service)
	servicesConn = make(map[net.Conn][]*Service)

	serve()

	// Will add internal "cellaserv" service setup here
}

// vim: set nowrap tw=100 noet sw=8:
