package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"net"
	"os"
	"time"
)

// http://wiki.wireshark.org/Development/LibpcapFileFormat
type PcapHeader struct {
	MagicNumber  uint32
	VersionMajor uint16
	VersionMinor uint16
	TimeZone     int32
	SigFigs      uint32
	SnapLen      uint32
	LinkType     uint32
}

type PacketHeader struct {
	Sec     uint32
	Usec    uint32
	InclLen uint32
	OrigLen uint32
}

// Not bufferized
var dumpFile *os.File

func dumpSetup() error {
	// do not use :=, it would shadow the global dumpFile variable
	var err error
	dumpFile, err = os.Create("cellaserv.dump")
	if err != nil {
		return err
	}

	// Write PCAP header
	header := PcapHeader{0xa1b2c3d4, 2, 4, 0, 0, 65535, 4200}
	err = binary.Write(dumpFile, binary.LittleEndian, header)

	return err
}

func dumpOutgoing(msg []byte) {
	sender := "cellaserv"
	logMsg := &cellaserv.LogMessage{Sender: &sender, Content: msg}
	dumpLogMessage(logMsg)
}

func dumpIncoming(conn net.Conn, msg []byte) {
	addr := conn.RemoteAddr().String()
	logMsg := &cellaserv.LogMessage{Sender: &addr, Content: msg}
	dumpLogMessage(logMsg)
}

func dumpLogMessage(msg *cellaserv.LogMessage) {
	msgBytes, _ := proto.Marshal(msg)

	// Write PCAP packet header
	now := time.Now()
	// Could use nanosecond PCAP format, but unsure of actual utility, and support by tools
	msgLen := uint32(len(msgBytes))
	header := PacketHeader{uint32(now.Unix()), uint32(now.Nanosecond() * 1000), msgLen, msgLen}
	binary.Write(dumpFile, binary.LittleEndian, header)

	// Write actual message
	dumpFile.Write(msgBytes)
}

// vim: set nowrap tw=100 noet sw=8:
