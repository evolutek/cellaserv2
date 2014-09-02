package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"container/list"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

var (
	// Command line flags
	versionFlag    = flag.Bool("version", false, "output version information and exit")
	sockPortFlag   = flag.String("port", "", "listening port")
	sockAddrListen = ":4200"

	// List of all currently handled connections
	connList *list.List

	// Map a connection to a name, filled with cellaserv.descrbie-conn
	connNameMap map[net.Conn]string

	// Map a connection to the service it spies
	connSpies map[net.Conn][]*Service

	// Map of currently connected services by name, then identification
	services map[string]map[string]*Service

	// Map of all services associated with a connection
	servicesConn map[net.Conn][]*Service

	// Map of requests ids with associated timeout timer
	reqIds             map[uint64]*RequestTracking
	subscriberMap      map[string][]net.Conn
	subscriberMatchMap map[string][]net.Conn
)

// Manage incoming connexions
func handle(conn net.Conn) {
	log.Info("[Net] Connection opened: %s", connDescribe(conn))

	connJson := connToJson(conn)
	cellaservPublish(logNewConnection, connJson)

	// Append to list of handled connections
	connListElt := connList.PushBack(conn)

	// Handle all messages received on this connection
	for {
		closed, err := handleMessage(conn)
		if err != nil {
			log.Error("[Message] %s", err)
		}
		if closed {
			log.Info("[Net] Connection closed: %s", connDescribe(conn))
			break
		}
	}

	// Remove from list of handled connection
	connList.Remove(connListElt)

	// Clean connection name, if not given this is a noop
	delete(connNameMap, conn)

	// Remove services registered by this connection
	// TODO: notify goroutines waiting for acks for this service
	for _, s := range servicesConn[conn] {
		log.Info("[Services] Remove %s", s)
		pub_json, _ := json.Marshal(s.JSONStruct())
		cellaservPublish(logLostService, pub_json)
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

					pub_json, _ := json.Marshal(
						LogSubscriberJSON{key, connDescribe(conn)})
					cellaservPublish(logLostSubscriber, pub_json)

					if len(subMap[key]) == 0 {
						delete(subMap, key)
						break
					}
				}
			}
		}
	}
	removeConnFromMap(subscriberMap)
	removeConnFromMap(subscriberMatchMap)

	// Remove spy info
	for _, srvc := range connSpies[conn] {
		for i, connItem := range srvc.Spies {
			if connItem == conn {
				// Remove from slice
				srvc.Spies[i] = srvc.Spies[len(srvc.Spies)-1]
				srvc.Spies = srvc.Spies[:len(srvc.Spies)-1]
				break
			}
		}
	}
	delete(connSpies, conn)

	cellaservPublish(logCloseConnection, connJson)
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

	if msgLen > 8*1024*1024 {
		return false, fmt.Errorf("Request too big: %d", msgLen)
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
		handleReply(conn, msgBytes, reply)
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
		handlePublish(conn, msgBytes, pub)
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

// Output version information and exit
func version() {
	fmt.Println("cellaserv2 version", csVersion)
	fmt.Println("Source: http://code.evolutek.org/cellaserv2")
	fmt.Println("Authors: ")
	fmt.Println("- Rémi Audebert")

	os.Exit(0)
}

func setup() {
	// Initialize our maps
	connNameMap = make(map[net.Conn]string)
	connSpies = make(map[net.Conn][]*Service)
	services = make(map[string]map[string]*Service)
	servicesConn = make(map[net.Conn][]*Service)
	reqIds = make(map[uint64]*RequestTracking)
	subscriberMap = make(map[string][]net.Conn)
	subscriberMatchMap = make(map[string][]net.Conn)
	connList = list.New()

	// Parse command line arguments
	flag.Parse()

	if *versionFlag {
		version()
	}

	// Setup basic logging facilities
	logPreSetup()

	settingsSetup()

	// Enable CPU profiling, stopped when cellaserv receive the kill request
	setupProfiling()

	// Setup cellaserv logging functions
	logSetup()

	// Setup pcap dumping of all packets
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
