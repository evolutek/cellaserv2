package main

import "net"

type Service struct {
	conn           net.Conn
	name           string
	identification string
}
