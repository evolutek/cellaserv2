package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"net"
)

func handlePublish(conn net.Conn, msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	log.Info("[Publish] %s publishes %s", conn.RemoteAddr(), *pub.Event)
	for _, pub := range subscriberMap[*pub.Event] {
		sendRawMessageLen(pub, msgLen, msgBytes)
	}
}

// vim: set nowrap tw=100 noet sw=8:
