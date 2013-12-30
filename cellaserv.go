package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"encoding/json"
	"net"
)

type ServiceJSON struct {
	Conn string
	Name string
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

func handleCellaservRequest(conn net.Conn, req *cellaserv.Request) {
	switch *req.Method {
	case "list-services":
		handleListServices(conn, req)
	case "list-connections":
		handleListConnections(conn, req)
	default:
		sendReplyError(conn, req, cellaserv.Reply_Error_NoSuchMethod)
	}
}

// vim: set nowrap tw=100 noet sw=8:
