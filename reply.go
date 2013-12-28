package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"net"
)

func handleReply(conn net.Conn, msgLen uint32, msgRaw []byte, rep *cellaserv.Reply) {
	id := *rep.Id
	log.Info("[Reply] %s replies to %d", conn.RemoteAddr(), id)

	sender, ok := reqIds[id]
	if !ok {
		log.Error("[Reply] Unknown ID: %d", id)
		return
	}
	delete(reqIds, id)

	log.Debug("[Reply] Forwarding to %s", sender.RemoteAddr())
	sendRawMessageLen(sender, msgLen, msgRaw)
}

// vim: set nowrap tw=100 noet sw=8:
