// main package
package main

import (
	handler "KeyValueStoreServer/server/handlers"
	log "KeyValueStoreServer/server/loggers"
	store "KeyValueStoreServer/server/store"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

const timeout = time.Second * 3

func main() {
	// fire up the loggers
	go log.WaitForAndProcessRequestLogs()
	go log.WaitForAndProcesslogs()

	port, depth := getcmdLine()

	// set store depth
	store.StoreDepth = depth

	log.InfoChannel <- fmt.Sprintf("Starting Server on %d", port)

	// create and monitor a new store
	go store.PublicAccess.Monitor()

	// register endpoint handlers
	setupHandlers()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: timeout,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errMessage := fmt.Sprintf("Error Starting Server %s", err)
			log.ErrorChannel <- errMessage

			os.Exit(-2)
		}
	}()

	<-handler.ShutdownServerChannel

	log.InfoChannel <- "Shutting Down Server"

	// close server
	_ = server.Close()

	// close loggers
	log.LoggerDoneChannel <- true
	log.RequestDoneChannel <- true
}

func getcmdLine() (int, int) {
	var port int

	var storeDepth int

	flag.IntVar(&port, "port", 0, "port to listen on")
	flag.IntVar(&storeDepth, "depth", 100, "max values to store default 100")
	flag.Parse()

	if port == 0 {
		log.ErrorChannel <- "No port parameter on command line"
		os.Exit(-1)
	}

	return port, storeDepth
}

func setupHandlers() {
	http.HandleFunc("/ping/", handler.ServePing)
	http.HandleFunc("/shutdown/", handler.ServeShutdown)
	http.HandleFunc(fmt.Sprintf("%s/", handler.BaseURLPath), handler.ServeKey)
	http.HandleFunc("/list/", handler.ServeList)
	http.HandleFunc("/login/", handler.ServeLogin)
}
