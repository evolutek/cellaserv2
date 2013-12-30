package main

import (
	"flag"
	"github.com/op/go-logging"
	golog "log"
	"os"
)

// Setup log
var log = logging.MustGetLogger("cellaserv")

var logRootDirectory = flag.String("log-root", ".", "root directory of logs")
var servicesLogs map[string]*golog.Logger

func logSetup() {
	servicesLogs = make(map[string]*golog.Logger)
}

func logSetupFile(what string) (l *golog.Logger) {
	l, ok := servicesLogs[what]
	if !ok {
		logFilename := *logRootDirectory + "/" + what + ".log"
		logFd, err := os.Create(logFilename)
		if err != nil {
			log.Error("[Log] Could not create log file:", logFilename)
			return
		}
		l = golog.New(logFd, what, golog.LstdFlags)
		l.SetPrefix("")
		servicesLogs[what] = l
	}
	return
}

func logEvent(event string, what string) {
	logger, ok := servicesLogs[event]
	if !ok {
		logger = logSetupFile(event)
	}
	logger.Println(what)
}

// vim: set nowrap tw=100 noet sw=8:
