package main

import (
	_ "bitbucket.org/evolutek/gocellaserv-protobuf"
	"github.com/op/go-logging"
	"net"
)

type Service struct {
	ip net.IPAddr
}

var log = logging.MustGetLogger("cellaserv")

func handle(conn net.Conn) {
	log.Debug("New connection")
}

func main() {
	ln, err := net.Listen("tcp", ":4200")
	if err != nil {
		log.Error("error")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go handle(conn)
	}
}
