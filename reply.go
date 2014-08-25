package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"net"
)

func handleReply(conn net.Conn, msgRaw []byte, rep *cellaserv.Reply) {
	id := *rep.Id
	log.Info("[Reply] %s replies to %d", conn.RemoteAddr(), id)

	reqTimer, ok := reqIds[id]
	if !ok {
		log.Error("[Reply] Unknown ID: %d", id)
		return
	}
	delete(reqIds, id)

	reqTimer.timer.Stop()
	log.Debug("[Reply] Forwarding to %s", reqTimer.sender.RemoteAddr())
	sendRawMessage(reqTrack.sender, msgRaw)
}

// vim: set nowrap tw=100 noet sw=8:
