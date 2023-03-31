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

const generalLogFile = "store.log"
const requestLogFile = "access.log"
const fileMode = 0600
const osAppend = 8
const osCreate = 512
const osWronly = 1

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

	file, err := os.OpenFile(requestLogFile, osAppend|osCreate|osWronly, fileMode)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = file.Close()
	}()

	infoLog.SetOutput(file)
	infoLog.Println(message)
}

// AddLog general logs to store.log file.
func addLog(logger *log.Logger, logEntry string) {
	file, err := os.OpenFile(generalLogFile, osAppend|osCreate|osWronly, fileMode)
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
func WaitForAndProcesslogs() {
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
func WaitForAndProcessRequestLogs() {
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
