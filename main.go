package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
)

// Command line flags
var sockPortFlag = flag.String("port", "", "listening port")
var sockAddrListen = ":4200"

// List of all currently handled connections
var connList []net.Conn

// Map of currently connected services by name, then identification
var services map[string]map[string]*Service

// Map of all services associated with a connection
var servicesConn map[net.Conn][]*Service

// Map of requests ids with associated timeout timer
var reqIds map[uint64]*RequestTimer
var subscriberMap map[string][]net.Conn
var subscriberMatchMap map[string][]net.Conn

// Internal log names
var logNewConnection = "log.new-connection"
var logCloseConnection = "log.close-connection"
var logLostService = "log.lost-service"

// Manage incoming connexions
func handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	log.Info("[Net] New %s", remoteAddr)
	cellaservPublish(logNewConnection, []byte(fmt.Sprintf("\"%s\"", remoteAddr)))

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

	log.Info("[Net] Connection closed: %s", remoteAddr)

	// Remove services registered by this connection
	// TODO: notify goroutines waiting for acks for this service
	// TODO: remove service from subscribers lists
	for _, s := range servicesConn[conn] {
		log.Info("[Services] Remove %s", s)
		pub, _ := json.Marshal(s.JSONStruct())
		cellaservPublish(logLostService, pub)
		delete(services[s.Name], s.Identification)
	}
	delete(servicesConn, conn)

	// Remove subscribes from this connection
	removeConnFromMap := func(subMap map[string][]net.Conn) {
		for key, subs := range subMap {
			for i, subConn := range subs {
				if conn == subConn {
					// Remove from list of subscribers
					subs[i] = subs[len(subs)-1]
					subMap[key] = subs[:len(subs)-1]
					if len(subs) == 0 {
						delete(subMap, key)
						break
					}
				}
			}
		}
	}
	removeConnFromMap(subscriberMap)
	removeConnFromMap(subscriberMatchMap)

	cellaservPublish(logCloseConnection, []byte(fmt.Sprintf("\"%s\"", remoteAddr)))
}

func logUnmarshalError(msg []byte) {
	dbg := ""
	for _, b := range msg {
		dbg = dbg + fmt.Sprintf("0x%02X ", b)
	}
	log.Error("[Net] Bad message: %s", dbg)
}

func handleMessage(conn net.Conn) (bool, error) {
	// Read message length as uint32
	var msgLen uint32
	err := binary.Read(conn, binary.BigEndian, &msgLen)
	if err != nil {
		return true, fmt.Errorf("Could not read message length: %s", err)
	}

	msgBytes := make([]byte, msgLen)
	_, err = conn.Read(msgBytes)
	if err != nil {
		return true, fmt.Errorf("Could not read message: %s", err)
	}

	// Dump raw msg to log
	dumpIncoming(conn, msgBytes)

	msg := &cellaserv.Message{}
	err = proto.Unmarshal(msgBytes, msg)
	if err != nil {
		logUnmarshalError(msgBytes)
		return false, fmt.Errorf("Could not unmarshal message: %s", err)
	}

	switch *msg.Type {
	case cellaserv.Message_Register:
		register := &cellaserv.Register{}
		err = proto.Unmarshal(msg.Content, register)
		if err != nil {
			logUnmarshalError(msg.Content)
			return false, fmt.Errorf("Could not unmarshal register: %s", err)
		}
		handleRegister(conn, register)
		return false, nil
	case cellaserv.Message_Request:
		request := &cellaserv.Request{}
		err = proto.Unmarshal(msg.Content, request)
		if err != nil {
			logUnmarshalError(msg.Content)
			return false, fmt.Errorf("Could not unmarshal request: %s", err)
		}
		handleRequest(conn, msgBytes, request)
		return false, nil
	case cellaserv.Message_Reply:
		reply := &cellaserv.Reply{}
		err = proto.Unmarshal(msg.Content, reply)
		if err != nil {
			logUnmarshalError(msg.Content)
			return false, fmt.Errorf("Could not unmarshal reply: %s", err)
		}
		handleReply(conn, msgLen, msgBytes, reply)
		return false, nil
	case cellaserv.Message_Subscribe:
		sub := &cellaserv.Subscribe{}
		err = proto.Unmarshal(msg.Content, sub)
		if err != nil {
			logUnmarshalError(msg.Content)
			return false, fmt.Errorf("Could not unmarshal subscribe: %s", err)
		}
		handleSubscribe(conn, sub)
		return false, nil
	case cellaserv.Message_Publish:
		pub := &cellaserv.Publish{}
		err = proto.Unmarshal(msg.Content, pub)
		if err != nil {
			logUnmarshalError(msg.Content)
			return false, fmt.Errorf("Could not unmarshal publish: %s", err)
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

func setup() {
	// Initialize our maps
	services = make(map[string]map[string]*Service)
	servicesConn = make(map[net.Conn][]*Service)
	reqIds = make(map[uint64]*RequestTimer)
	subscriberMap = make(map[string][]net.Conn)
	subscriberMatchMap = make(map[string][]net.Conn)

	// Parse command line arguments
	flag.Parse()

	// Setup cellaserv log
	logPreSetup()

	settingsSetup()

	setupProfiling()

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
