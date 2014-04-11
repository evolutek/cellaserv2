package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"net"
	"strings"
)

func handlePublish(conn net.Conn, msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	log.Info("[Publish] %s publishes %s", conn.RemoteAddr(), *pub.Event)
	doPublish(msgLen, msgBytes, pub)
}

func doPublish(msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	event := pub.Event
	if strings.HasPrefix(*event, "log.") {
		cellaservLog(pub)
	}

	for _, pub := range subscriberMap[*event] {
		log.Debug("[Publish] Forwarding publish to %s", pub.RemoteAddr())
		dumpOutgoing(msgBytes)
		sendRawMessageLen(pub, msgLen, msgBytes)
	}
}

// vim: set nowrap tw=100 noet sw=8:
