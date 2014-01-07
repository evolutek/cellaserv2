package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime/pprof"
)

var sockPortFlag = flag.String("port", "", "listening port")
var sockAddrListen = ":4200"

// Currently connected services
var services map[string]map[string]*Service
var servicesConn map[net.Conn][]*Service
var reqIds map[uint64]*RequestTimer
var subscriberMap map[string][]net.Conn

// Manage incoming connexions
func handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	log.Info("[Net] New %s", remoteAddr)

	// Handle all messages received on this connection
	for {
		closed, err := handleMessage(conn)
		if closed {
			break
		}
		if err != nil {
			log.Error("[Message] %s", err)
		}
	}

	// Remove services registered by this connection
	// TODO: notify goroutines waiting for acks for this service
	for _, s := range servicesConn[conn] {
		log.Info("[Services] Remove %s", s)
		delete(services[s.Name], s.Identification)
	}

	log.Info("[Net] Connection closed: %s", remoteAddr)
}

func handleMessage(conn net.Conn) (bool, error) {
	// Read message length as uint32
	var msgLen uint32
	err := binary.Read(conn, binary.BigEndian, &msgLen)
	if err != nil {
		return true, fmt.Errorf("Could not read message length:", err)
	}

	msgBytes := make([]byte, msgLen)
	_, err = conn.Read(msgBytes)
	if err != nil {
		return true, fmt.Errorf("Could not read message:", err)
	}

	// Dump raw msg to log
	dumpIncoming(conn, msgBytes)

	msg := &cellaserv.Message{}
	err = proto.Unmarshal(msgBytes, msg)
	if err != nil {
		return false, fmt.Errorf("Could not unmarshal message:", err)
	}

	switch *msg.Type {
	case cellaserv.Message_Register:
		register := &cellaserv.Register{}
		err = proto.Unmarshal(msg.Content, register)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal register:", err)
		}
		handleRegister(conn, register)
		return false, nil
	case cellaserv.Message_Request:
		request := &cellaserv.Request{}
		err = proto.Unmarshal(msg.Content, request)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal request:", err)
		}
		handleRequest(conn, msgBytes, request)
		return false, nil
	case cellaserv.Message_Reply:
		reply := &cellaserv.Reply{}
		err = proto.Unmarshal(msg.Content, reply)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal reply:", err)
		}
		handleReply(conn, msgLen, msgBytes, reply)
		return false, nil
	case cellaserv.Message_Subscribe:
		sub := &cellaserv.Subscribe{}
		err = proto.Unmarshal(msg.Content, sub)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal subscribe:", err)
		}
		handleSubscribe(conn, sub)
		return false, nil
	case cellaserv.Message_Publish:
		pub := &cellaserv.Publish{}
		err = proto.Unmarshal(msg.Content, pub)
		if err != nil {
			return false, fmt.Errorf("Could not unmarshal publish:", err)
		}
		handlePublish(conn, msgLen, msgBytes, pub)
		return false, nil
	default:
		return false, fmt.Errorf("Unknown message type: %d", *msg.Type)
	}
}

// Start listening and receiving connections
func serve() {
	ln, err := net.Listen("tcp", sockAddrListen)
	if err != nil {
		log.Error("[Net] Could not listen: %s", err)
		return
	}
	defer ln.Close()

	log.Info("[Net] Listening on %s", sockAddrListen)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("[Net] Could not accept: %s", err)
			continue
		}

		go handle(conn)
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")

func setup() {
	// Initialize our maps
	services = make(map[string]map[string]*Service)
	servicesConn = make(map[net.Conn][]*Service)
	reqIds = make(map[uint64]*RequestTimer)
	subscriberMap = make(map[string][]net.Conn)

	logPreSetup()

	// Parse arguments
	flag.Parse()

	settingsSetup()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	logSetup()

	// Setup dumping
	err := dumpSetup()
	if err != nil {
		log.Error("Could not setup dump: %s", err)
	}
}

func main() {
	setup()
	serve()
}

// vim: set nowrap tw=100 noet sw=8:
