package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"net"
)

func handleSubscribe(conn net.Conn, sub *cellaserv.Subscribe) {
	log.Info("[Subscribe] %s subscribes to %s", conn.RemoteAddr(), *sub.Event)
	subscriberMap[*sub.Event] = append(subscriberMap[*sub.Event], conn)
}

// vim: set nowrap tw=100 noet sw=8:
