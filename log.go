package main

import (
	"flag"
	"github.com/op/go-logging"
	golog "log"
	"os"
)

// Setup log
var log *logging.Logger

var logLevel = logging.WARNING
var logRootDirectory = flag.String("log-root", ".", "root directory of logs")
var logLevelFlag = flag.String("log-level", "", "logger verbosity")

var servicesLogs map[string]*golog.Logger

// Setup that must be done before any log is made
func logPreSetup() {
	format := logging.MustStringFormatter("%{level:-7s} %{time:Jan _2 15:04:05.000} %{message}")

	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)

	logging.SetFormatter(format)
	log = logging.MustGetLogger("cellaserv")
}

func logSetup() {
	logging.SetLevel(logLevel, "cellaserv")
	servicesLogs = make(map[string]*golog.Logger)
}

func logSetupFile(what string) (l *golog.Logger) {
	l, ok := servicesLogs[what]
	if !ok {
		logFilename := *logRootDirectory + "/" + what + ".log"
		logFd, err := os.Create(logFilename)
		if err != nil {
			log.Error("[Log] Could not create log file: %s", logFilename)
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
		if logger == nil {
			return
		}
	}
	logger.Println(what)
}

// vim: set nowrap tw=100 noet sw=8:
