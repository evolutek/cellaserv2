package main

import (
	"encoding/binary"
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

func dumpMessage(msg []byte) {
	// Write PCAP packet header
	now := time.Now()
	// Could use nanosecond PCAP format, but unsure of actual use, and support by other tools
	msgLen := uint32(len(msg))
	header := PacketHeader{uint32(now.Unix()), uint32(now.Nanosecond() * 1000), msgLen, msgLen}
	binary.Write(dumpFile, binary.LittleEndian, header)

	// Write actual message
	dumpFile.Write(msg)
}

// vim: set nowrap tw=100 noet sw=8:
