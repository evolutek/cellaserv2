package main

import (
	"flag"
	"github.com/op/go-logging"
	golog "log"
	"os"
)

// log is the main cellaserv logger, use it everywhere you want!
var log *logging.Logger

// Default log level
var logLevel = logging.WARNING

// Command line flags
var logRootDirectory = flag.String("log-root", ".", "root directory of logs")
var logLevelFlag = flag.String("log-level", "", "logger verbosity")
var logToFile = flag.String("log-file", "", "log to custom file instead of stderr")

// Map of the logger associated with a service
var servicesLogs map[string]*golog.Logger

// Setup that must be done before any log is made. Command line arguments parsing must be done
// before calling logPreSetup()
func logPreSetup() {
	format := logging.MustStringFormatter("%{level:-7s} %{time:Jan _2 15:04:05.000} %{message}")

	var logBackend *logging.LogBackend
	if *logToFile != "" {
		// Use has specified a log file to use
		logFile, err := os.OpenFile(*logToFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			golog.Println(err)
			golog.Println("Falling back on log on stderr")
			logBackend = logging.NewLogBackend(os.Stderr, "", 0)
		} else {
			logBackend = logging.NewLogBackend(logFile, "", 0)
		}
	} else {
		// Log on stderr
		logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	}
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
		logFd, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
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
