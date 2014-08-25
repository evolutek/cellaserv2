package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// This string will be replaced by the build system
var csVersion = "git"

// Logs sent by cellaserv
var logCloseConnection = "log.cellaserv.close-connection"
var logConnRename = "log.cellaserv.connection-rename"
var logLostService = "log.cellaserv.lost-service"
var logLostSubscriber = "log.cellaserv.lost-subscriber"
var logNewConnection = "log.cellaserv.new-connection"
var logNewService = "log.cellaserv.new-service"
var logNewSubscriber = "log.cellaserv.new-subscriber"

type connDescribeRequest struct {
	Name string
}

type connNameJSON struct {
	Addr string
	Name string
}

/*
handleDescribeConn attaches a name to the connection that sent the request.

This information is normaly given when a service registers, but it can also be useful for other
clients.
*/
func handleDescribeConn(conn net.Conn, req *cellaserv.Request) {
	var data connDescribeRequest

	if err := json.Unmarshal(req.Data, &data); err != nil {
		log.Warning("[Cellaserv] Could not unmarshal describe-conn: %s, %s", req.Data, err)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	connNameMap[conn] = data.Name
	newName := connDescribe(conn)

	pub_json, _ := json.Marshal(connNameJSON{conn.RemoteAddr().String(), newName})
	cellaservPublish(logConnRename, pub_json)

	log.Debug("[Cellaserv] Describe %s as %s", conn.RemoteAddr(), data.Name)

	sendReply(conn, req, nil) // Empty reply
}

func handleListServices(conn net.Conn, req *cellaserv.Request) {
	// Fix static empty slice that is "null" in JSON
	// A dynamic empty slice is []
	servicesList := make([]*ServiceJSON, 0)
	for _, names := range services {
		for _, s := range names {
			servicesList = append(servicesList, s.JSONStruct())
		}
	}

	data, err := json.Marshal(servicesList)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal the services")
	}
	sendReply(conn, req, data)
}

// handleListConnections replies with the list of currently connected clients
func handleListConnections(conn net.Conn, req *cellaserv.Request) {
	var conns []connNameJSON
	for c := connList.Front(); c != nil; c = c.Next() {
		connElt := c.Value.(net.Conn)
		conns = append(conns,
			connNameJSON{connElt.RemoteAddr().String(), connDescribe(connElt)})
	}
	data, err := json.Marshal(conns)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal the connections list")
	}
	sendReply(conn, req, data)
}

// handleListEvents replies with the list of subscribers
func handleListEvents(conn net.Conn, req *cellaserv.Request) {
	events := make(map[string][]string)

	fillMap := func(subMap map[string][]net.Conn) {
		for event, conns := range subMap {
			var connSlice []string
			for _, connItem := range conns {
				connSlice = append(connSlice, connItem.RemoteAddr().String())
			}
			events[event] = connSlice
		}
	}

	fillMap(subscriberMap)
	fillMap(subscriberMatchMap)

	data, err := json.Marshal(events)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal the event list")
	}
	sendReply(conn, req, data)
}

/*
handleGetLogs reply with the content of log files.

Request format:

	bytes

Examples:

	cellaserv.new-connection
	cellaserv.*

Reply format:

	bytes

*/
func handleGetLogs(conn net.Conn, req *cellaserv.Request) {
	if req.Data == nil {
		log.Warning("[Cellaserv] Log request does not specify event")
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	event := string(req.Data)
	pattern := path.Join(*logRootDirectory, logSubDir, event+".log")

	if !strings.HasPrefix(pattern, path.Join(*logRootDirectory, logSubDir)) {
		log.Warning("[Cellaserv] Don't try to do directory traversal")
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	// Globbing is allowed
	filenames, err := filepath.Glob(pattern)

	if err != nil {
		log.Warning("[Cellaserv] Invalid log globbing : %s, %s", event, err)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	if len(filenames) == 0 {
		log.Warning("[Cellaserv] No such logs: %s", event)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	var data_buf bytes.Buffer

	for _, filename := range filenames {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Warning("[Cellaserv] Could not open log: %s", filename)
			sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
			return
		}
		data_buf.Write(data)
	}
	sendReply(conn, req, data_buf.Bytes())
}

func handleLogRotate(conn net.Conn, req *cellaserv.Request) {
	type logRotateFormat struct {
		Where string
	}
	// Default to time
	if req.Data == nil {
		logRotateTimeNow()
	} else {
		var data logRotateFormat
		err := json.Unmarshal(req.Data, &data)
		if err != nil {
			log.Warning("[Cellaserv] Could not rotate log, json error: %s", err)
			sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
			return
		}
		logRotateName(data.Where)
	}
	sendReply(conn, req, nil)
}

// handleShutdown quit cellaserv. Used for debug purposes
func handleShutdown() {
	stopProfiling()
	os.Exit(0)
}

// handleVersion return the version of cellaserv
func handleVersion(conn net.Conn, req *cellaserv.Request) {
	data, err := json.Marshal(csVersion)
	if err != nil {
		log.Warning("[Cellaserv] Could not marshall version, json error: %s", err)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}
	sendReply(conn, req, data)
}

func cellaservRequest(conn net.Conn, req *cellaserv.Request) {
	switch *req.Method {
	case "describe-conn", "describe_conn":
		handleDescribeConn(conn, req)
	case "get-logs", "get_logs":
		handleGetLogs(conn, req)
	case "list-connections", "list_connections":
		handleListConnections(conn, req)
	case "list-events", "list_events":
		handleListEvents(conn, req)
	case "list-services", "list_services":
		handleListServices(conn, req)
	case "log-rotate", "log_rotate":
		handleLogRotate(conn, req)
	case "shutdown":
		handleShutdown()
	case "version":
		handleVersion(conn, req)
	default:
		sendReplyError(conn, req, cellaserv.Reply_Error_NoSuchMethod)
	}
}

func cellaservLog(pub *cellaserv.Publish) {
	if pub.Data == nil {
		log.Warning("[Log] %s does not have data", *pub.Event)
		return
	}
	data := string(pub.Data)
	event := (*pub.Event)[4:]
	logEvent(event, data)
}

func cellaservPublish(event string, data []byte) {
	pub := &cellaserv.Publish{Event: &event}
	if data != nil {
		pub.Data = data
	}
	pubBytes, err := proto.Marshal(pub)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal event")
		return
	}
	msgType := cellaserv.Message_Publish
	msg := &cellaserv.Message{Type: &msgType, Content: pubBytes}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal event")
		return
	}

	doPublish(uint32(len(msgBytes)), msgBytes, pub)
}

// vim: set nowrap tw=100 noet sw=8:
