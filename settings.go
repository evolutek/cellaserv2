package main

import (
	"code.google.com/p/gcfg"
	"github.com/op/go-logging"
	"os"
)

var cfg struct {
	Cellaserv struct {
		Debug string
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
		log.Warning("[Config] Unknown debug value: %s", cfg.Cellaserv.Debug)
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
		log.Debug("[Config] Could not read or parse '/etc/conf.d/cellaserv'",
			err)
		return
	}

	// Override if not ""
	setLogLevelFromString(cfg.Cellaserv.Debug)
	setLogLevelFromString(os.Getenv("CS_DEBUG"))
	setLogLevelFromString(*logLevelFlag)

	setSockAddrListenFromString(":" + cfg.Cellaserv.Port)
	setSockAddrListenFromString(":" + os.Getenv("CS_PORT"))
	setSockAddrListenFromString(":" + *sockPortFlag)
}
