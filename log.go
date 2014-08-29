package main

import (
	"encoding/json"
	"flag"
	"github.com/op/go-logging"
	golog "log"
	"os"
	"path"
	"time"
)

// log is the main cellaserv logger, use it everywhere you want!
var log *logging.Logger

// Default log level
var logLevel = logging.WARNING

// Command line flags
var logRootDirectory = flag.String("log-root", ".", "root directory of logs")
var logSubDir string
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
	// Set default log subDirectory to now
	logRotateTimeNow()
}

// logRotateName set the new log subdirectory to name
func logRotateName(name string) {
	log.Debug("[Log] Rotating to \"%s\"", name)
	logSubDir = name
	logFullDir := path.Join(*logRootDirectory, logSubDir)
	err := os.MkdirAll(logFullDir, 0755)
	if err != nil {
		log.Error("[Log] Could not create log directories, %s: %s", logFullDir, err)
	}
	// XXX: close old log files?
	servicesLogs = make(map[string]*golog.Logger)

	pub_data, err := json.Marshal(logSubDir)
	if err != nil {
		log.Error("[Publish] Could not publish new log session, json error: %s: %s",
			logSubDir, err)
	}
	cellaservPublish(logNewLogSession, pub_data)
}

// logRotateTimeNow switch the current log subdirectory to current time
func logRotateTimeNow() {
	now := time.Now()
	newSubDir := now.Format(time.Stamp)
	logRotateName(newSubDir)
}

func logSetupFile(what string) (l *golog.Logger) {
	l, ok := servicesLogs[what]
	if !ok {
		logFilename := path.Join(*logRootDirectory, logSubDir, what+".log")
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
