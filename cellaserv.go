package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"github.com/golang/protobuf/proto"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// This string will be replaced by the build system
	csVersion = "git"

	// Logs sent by cellaserv
	logCloseConnection = "log.cellaserv.close-connection"
	logConnRename      = "log.cellaserv.connection-rename"
	logLostService     = "log.cellaserv.lost-service"
	logLostSubscriber  = "log.cellaserv.lost-subscriber"
	logNewConnection   = "log.cellaserv.new-connection"
	logNewService      = "log.cellaserv.new-service"
	logNewSubscriber   = "log.cellaserv.new-subscriber"
	logNewLogSession   = "log.cellaserv.new-log-session"
)

// Send conn data as this struct
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
	var data struct {
		Name string
	}

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
handleGetLogs replies with the content of log files.

Request format:

	bytes

Examples:

	cellaserv.new-connection
	cellaserv.*

Reply format:

	map[string]string

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

	logs := make(map[string]string)

	for _, filename := range filenames {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Warning("[Cellaserv] Could not open log: %s", filename)
			sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
			return
		}
		logs[filename] = string(data)
	}

	logs_json, _ := json.Marshal(logs)
	sendReply(conn, req, logs_json)
}

// handleLogRotate changes the current log environment
func handleLogRotate(conn net.Conn, req *cellaserv.Request) {
	// Default to time
	if req.Data == nil {
		logRotateTimeNow()
	} else {
		var data struct {
			Where string
		}

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

// handleSession returns the current log sesion
func handleSession(conn net.Conn, req *cellaserv.Request) {
	data, err := json.Marshal(logSubDir)
	if err != nil {
		log.Warning("[Cellaserv] Could not marshall log session, json error: %s", err)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}
	sendReply(conn, req, data)
}

// handleShutdown quits cellaserv. Used for debug purposes
func handleShutdown() {
	stopProfiling()
	os.Exit(0)
}

// handleSpy registers the connection as a spy of a service
func handleSpy(conn net.Conn, req *cellaserv.Request) {
	var data struct {
		Service        string
		Identification string
	}

	err := json.Unmarshal(req.Data, &data)
	if err != nil {
		log.Warning("[Cellaserv] Could not spy, json error: %s", err)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	srvc, ok := services[data.Service][data.Identification]
	if !ok {
		log.Warning("[Cellaserv] Could not spy, no such service: %s %s", data.Service,
			data.Identification)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}

	log.Debug("[Cellaserv] %s spies on %s/%s", connDescribe(conn), data.Service,
		data.Identification)

	srvc.Spies = append(srvc.Spies, conn)
	connSpies[conn] = append(connSpies[conn], srvc)

	sendReply(conn, req, nil)
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
	case "session":
		handleSession(conn, req)
	case "shutdown":
		handleShutdown()
	case "spy":
		handleSpy(conn, req)
	case "version":
		handleVersion(conn, req)
	default:
		sendReplyError(conn, req, cellaserv.Reply_Error_NoSuchMethod)
	}
}

// cellaservLog logs a publish message to a file
func cellaservLog(pub *cellaserv.Publish) {
	var data string
	if pub.Data != nil {
		data = string(pub.Data)
	}
	event := (*pub.Event)[4:] // Strip 'log.'
	logEvent(event, data)
}

// cellaservPublish sends a publish message from cellaserv
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

	doPublish(msgBytes, pub)
}

// vim: set nowrap tw=100 noet sw=8:
