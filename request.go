package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"net"
)

func replyError(conn net.Conn, id *uint64, err_t cellaserv.Reply_Error_Type) {
	err := &cellaserv.Reply_Error{Type: &err_t}

	reply := &cellaserv.Reply{Error: err}
	replyBytes, _ := proto.Marshal(reply)

	msgType := cellaserv.Message_Reply
	msg := &cellaserv.Message{
		Type:    &msgType,
		Content: replyBytes,
	}
	sendMessage(conn, msg)
}

func handleRequest(conn net.Conn, msgLen uint32, msgRaw []byte, req *cellaserv.Request) {
	log.Info("[Request] New request from %s", conn.RemoteAddr())

	// Checks from Get*() methods are useless
	name := req.ServiceName
	var ident *string
	if req.ServiceIdentification != nil {
		ident = req.ServiceIdentification
	}
	method := req.Method
	id := req.Id
	log.Debug("[Request] id:%s %s[%s].%s", id, name, ident, method)

	idents, ok := services[*name]
	if !ok {
		log.Warning("[Request] id:%s No such service: %s", *id, *name)
		replyError(conn, id, cellaserv.Reply_Error_NoSuchService)
		return
	}
	var srvc *Service
	if ident != nil {
		srvc, ok = idents[*ident]
		if !ok {
			log.Warning("[Request] id:%s No such identification for service %s: %s",
				id, *name, *ident)
		}
	} else {
		srvc, ok = idents[""]
		if !ok {
			log.Warning("[Request] id:%s Must use identification for service %s",
				id, *name)
		}
	}
	if !ok {
		replyError(conn, id, cellaserv.Reply_Error_InvalidIdentification)
		return
	}

	log.Debug("[Request] Forwarding request to %s", srvc)
	sendRawMessageLen(srvc.conn, msgLen, msgRaw)
}

// vim: set nowrap tw=100 noet sw=8:
