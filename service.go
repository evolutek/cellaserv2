package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Service struct {
	Conn           net.Conn
	Name           string
	Identification string
	// Internal buffer used to craft messages
	buf bytes.Buffer
}

func newService(conn net.Conn, name string, ident string) *Service {
	var buf bytes.Buffer
	s := &Service{conn, name, ident, buf}
	return s
}

func (s *Service) String() string {
	return fmt.Sprintf("{Service %s[%s] at %s}", s.Name, s.Identification, s.Conn.RemoteAddr())
}

func (s *Service) sendMessage(msg []byte) {
	s.buf.Reset()
	binary.Write(&s.buf, binary.BigEndian, uint32(len(msg)))
	s.buf.Write(msg)
	s.Conn.Write(s.buf.Bytes())
}

// vim: set nowrap tw=100 noet sw=8:
