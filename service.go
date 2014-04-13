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

type ServiceJSON struct {
	Conn           string
	Name           string
	Identification string
}

func newService(conn net.Conn, name string, ident string) *Service {
	var buf bytes.Buffer
	s := &Service{conn, name, ident, buf}
	return s
}

func (s *Service) String() string {
	return fmt.Sprintf("{Service %s/%s at %s}", s.Name, s.Identification, s.Conn.RemoteAddr())
}

// JSONStruct creates a struc good for JSON encoding.
func (s *Service) JSONStruct() *ServiceJSON {
	return &ServiceJSON{
		Conn:           s.Conn.RemoteAddr().String(),
		Name:           s.Name,
		Identification: s.Identification,
	}
}

func (s *Service) sendMessage(msg []byte) {
	s.buf.Reset()
	// Write the size of the message
	binary.Write(&s.buf, binary.BigEndian, uint32(len(msg)))
	// Concatenate with message content
	s.buf.Write(msg)
	// Send the whole message at once
	s.Conn.Write(s.buf.Bytes())
}

// vim: set nowrap tw=100 noet sw=8:
