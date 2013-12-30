package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"net"
	"time"
)

type RequestTimer struct {
	sender net.Conn
	timer  *time.Timer
}

func handleRequest(conn net.Conn, msgLen uint32, msgRaw []byte, req *cellaserv.Request) {
	log.Info("[Request] Incoming from %s", conn.RemoteAddr())

	// Runtime checks in Get*() functions are useless
	name := req.ServiceName
	method := req.Method
	id := req.Id

	var ident *string
	if req.ServiceIdentification != nil {
		ident = req.ServiceIdentification
		log.Debug("[Request] id:%d %s[%s].%s", *id, *name, *ident, *method)
	} else {
		log.Debug("[Request] id:%d %s.%s", *id, *name, *method)
	}

	if *name == "cellaserv" {
		cellaservRequest(conn, req)
		return
	}

	idents, ok := services[*name]
	if !ok || len(idents) == 0 {
		log.Warning("[Request] id:%d No such service: %s", *id, *name)
		sendReplyError(conn, req, cellaserv.Reply_Error_NoSuchService)
		return
	}
	var srvc *Service
	if ident != nil {
		srvc, ok = idents[*ident]
		if !ok {
			log.Warning("[Request] id:%d No such identification for service %s: %s",
				*id, *name, *ident)
		}
	} else {
		srvc, ok = idents[""]
		if !ok {
			log.Warning("[Request] id:%d Must use identification for service %s",
				*id, *name)
		}
	}
	if !ok {
		sendReplyError(conn, req, cellaserv.Reply_Error_InvalidIdentification)
		return
	}

	// Handle timeouts
	handleTimeout := func() {
		_, ok := reqIds[*id]
		if ok {
			log.Error("[Request] id:%d Timeout of %s", *id, srvc)
			sendReplyError(conn, req, cellaserv.Reply_Error_Timeout)
		}
	}
	timer := time.AfterFunc(5*time.Second, handleTimeout)

	// The ID is used to track the sender of the request
	reqIds[*id] = &RequestTimer{conn, timer}

	log.Debug("[Request] Forwarding to %s", srvc)
	sendRawMessageLen(srvc.Conn, msgLen, msgRaw)
}

// vim: set nowrap tw=100 noet sw=8:
