package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"net"
	"strings"
)

func handlePublish(conn net.Conn, msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	event := pub.Event
	log.Info("[Publish] %s publishes %s", conn.RemoteAddr(), *event)

	if strings.HasPrefix(*event, "log.") {
		cellaservLog(pub)
	}

	for _, pub := range subscriberMap[*event] {
		sendRawMessageLen(pub, msgLen, msgBytes)
	}
}

// vim: set nowrap tw=100 noet sw=8:
