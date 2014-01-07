package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"runtime/pprof"
)

type ServiceJSON struct {
	Conn           string
	Name           string
	Identification string
}

func handleListServices(conn net.Conn, req *cellaserv.Request) {
	var servicesList []*ServiceJSON
	for _, names := range services {
		for _, s := range names {
			servicesList = append(servicesList, &ServiceJSON{
				s.Conn.RemoteAddr().String(),
				s.Name,
				s.Identification,
			})
		}
	}

	data, err := json.Marshal(servicesList)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal the services")
	}
	sendReply(conn, req, data)
}

func handleListConnections(conn net.Conn, req *cellaserv.Request) {
	var conns []string
	for c := range servicesConn {
		conns = append(conns, c.RemoteAddr().String())
	}
	data, err := json.Marshal(conns)
	if err != nil {
		log.Error("[Cellaserv] Could not marshal the connecions")
	}
	sendReply(conn, req, data)
}

func handleGetLogs(conn net.Conn, req *cellaserv.Request) {
	if req.Data == nil {
		log.Warning("[Cellaserv] Log request does not specify event")
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
		return
	}
	event := string(req.Data)
	filename := *logRootDirectory + "/" + event + ".log"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warning("[Cellaserv] Could not open log:", filename)
		sendReplyError(conn, req, cellaserv.Reply_Error_BadArguments)
	}
	sendReply(conn, req, data)
}

func handleShutdown() {
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	os.Exit(0)
}

func cellaservRequest(conn net.Conn, req *cellaserv.Request) {
	switch *req.Method {
	case "list-services":
		handleListServices(conn, req)
	case "list-connections":
		handleListConnections(conn, req)
	case "get-log":
		handleGetLogs(conn, req)
	case "shutdown":
		handleShutdown()
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

// vim: set nowrap tw=100 noet sw=8:
