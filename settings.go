package main

import (
	"os"

	"gopkg.in/gcfg.v1"
	"github.com/op/go-logging"
)

var cfg struct {
	Cellaserv struct {
		Debug string
		Port  string
	}
	Client struct {
		Debug string
		Host  string
		Port  string
	}
}

func setLogLevelFromString(lvl string) {
	switch lvl {
	case "":
	case "0":
		logLevel = logging.WARNING
	case "1":
		logLevel = logging.INFO
	case "2":
		logLevel = logging.DEBUG
	default:
		log.Warning("[Config] Unknown debug value: %s", lvl)
	}
}

func setSockAddrListenFromString(addr string) {
	if addr != "" && addr != ":" {
		sockAddrListen = addr
	}
}

func settingsSetup() {
	err := gcfg.ReadFileInto(&cfg, "/etc/conf.d/cellaserv")
	if err != nil {
		// Not fatal, all values will be ""
		log.Debug("[Config] %s", err)
	}

	// Override if not ""
	setLogLevelFromString(cfg.Cellaserv.Debug)
	setLogLevelFromString(os.Getenv("CS_DEBUG"))
	setLogLevelFromString(*logLevelFlag)

	setSockAddrListenFromString(":" + cfg.Cellaserv.Port)
	setSockAddrListenFromString(":" + os.Getenv("CS_PORT"))
	setSockAddrListenFromString(":" + *sockPortFlag)
}
