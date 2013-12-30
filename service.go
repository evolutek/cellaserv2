package main

import (
	"fmt"
	"net"
)

type Service struct {
	Conn           net.Conn
	Name           string
	Identification string
}

func (s *Service) String() string {
	return fmt.Sprintf("{Service %s[%s] at %s}", s.Name, s.Identification, s.Conn.RemoteAddr())
}

// vim: set nowrap tw=100 noet sw=8:
