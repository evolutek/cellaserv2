package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"net"
	"path/filepath"
	"strings"
)

func handlePublish(conn net.Conn, msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	log.Info("[Publish] %s publishes %s", connDescribe(conn), *pub.Event)
	doPublish(msgLen, msgBytes, pub)
}

func doPublish(msgLen uint32, msgBytes []byte, pub *cellaserv.Publish) {
	event := *pub.Event

	// Handle log publishes
	if strings.HasPrefix(event, "log.") {
		cellaservLog(pub)
	}

	var subs []net.Conn

	// Handle glob susbscribers
	for pattern, cons := range subscriberMatchMap {
		matched, _ := filepath.Match(pattern, event)
		if matched {
			subs = append(subs, cons...)
		}
	}

	// Add exact matches
	subs = append(subs, subscriberMap[event]...)

	for _, connSub := range subs {
		log.Debug("[Publish] Forwarding publish to %s", connDescribe(connSub))
		dumpOutgoing(connSub, msgBytes)
		sendRawMessageLen(connSub, msgLen, msgBytes)
	}
}

// vim: set nowrap tw=100 noet sw=8:
