// Package loggers various channels for logging
package loggers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	infoLog  = log.New(nil, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLog  = log.New(nil, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(nil, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

var generalLogFile = "store.log"
var requestLogFile = "access.log"

// RequestChannel for logging requests.
var RequestChannel = make(chan *http.Request)

// InfoChannel for logging information.
var InfoChannel = make(chan string)

// WarnChannel for logging warnings.
var WarnChannel = make(chan string)

// ErrorChannel for logging errors.
var ErrorChannel = make(chan string)

// RequestDoneChannel for letting us know request logging has stopped.
var RequestDoneChannel = make(chan bool)

// LoggerDoneChannel for letting us know general logging has stopped.
var LoggerDoneChannel = make(chan bool)

// addRequestlog logs request information to the access.log file.
func addRequestlog(r *http.Request) {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf("Request received %s, %s, %s, %s", currentTime, r.Method, r.URL.Path, r.RemoteAddr)
	writelog(infoLog, message, requestLogFile)
}

// AddLog general logs to store.log file.
func addLog(logger *log.Logger, logEntry string) {
	writelog(logger, logEntry, generalLogFile)
}

// writeLog writes logging information to the specfied file.
func writelog(logger *log.Logger, logEntry, logFile string) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = file.Close()
	}()

	logger.SetOutput(file)
	logger.Println(logEntry)
}

// WaitForAndProcesslogs waits on channels for general logs and writes to gerneral log file.
// pass "" for logile to use default store.log or specify required
func WaitForAndProcesslogs(logfile string) {
	if logfile != "" {
		generalLogFile = logfile
	}

	loop := true

	for loop {
		select {
		case i := <-InfoChannel:
			fmt.Println(i)
			addLog(infoLog, i)
		case w := <-WarnChannel:
			addLog(warnLog, w)
		case e := <-ErrorChannel:
			addLog(errorLog, e)
		case r := <-RequestChannel:
			addRequestlog(r)
		case <-LoggerDoneChannel:
			defer func() {
				close(InfoChannel)
				close(WarnChannel)
				close(ErrorChannel)
				close(LoggerDoneChannel)
			}()

			loop = false
		}
	}
}

// WaitForAndProcessRequestLogs waits on channel for request logs and writes to request log file.
// pass "" for logile to use default access.log or specify required
func WaitForAndProcessRequestLogs(logfile string) {
	if logfile != "" {
		requestLogFile = logfile
	}

	loop := true

	for loop {
		select {
		case r := <-RequestChannel:
			addRequestlog(r)
		case <-RequestDoneChannel:
			defer func() {
				close(RequestChannel)
				close(RequestDoneChannel)
			}()

			loop = false
		}
	}
}
