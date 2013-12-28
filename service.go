package main

import (
	"fmt"
	"net"
)

type Service struct {
	conn           net.Conn
	name           string
	identification string
}

func (s *Service) String() string {
	return fmt.Sprintf("{Service %s[%s] at %s}", s.name, s.identification, s.conn.RemoteAddr())
}

// vim: set nowrap tw=100 noet sw=8:
