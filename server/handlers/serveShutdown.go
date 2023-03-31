// Package handlers handle shutdown.
package handlers

import (
	log "KeyValueStoreServer/server/loggers"
	"fmt"
	"net/http"
)

// ShutdownServerChannel channel to monitor for shutting down the server.
var ShutdownServerChannel = make(chan int)

// ServeShutdown checks to see if call has authorisation to shutdown
// if so puts entry on shutdown channel.
func ServeShutdown(writer http.ResponseWriter, req *http.Request) {
	log.RequestChannel <- req

	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writer.Header().Set("Content-Type", "text/plain charset=utf-8")

	if username := getUsername(req); username == Admin {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("OK"))

		ShutdownServerChannel <- 1

		return
	}
	log.WarnChannel <- "Attempt to shutdown without supplying authorisation"

	writer.WriteHeader(http.StatusForbidden)
	_, err := writer.Write([]uint8("Forbidden"))

	if err != nil {
		warnMessage := fmt.Sprintf("Could not write to response %v", err)
		log.WarnChannel <- warnMessage

		writer.WriteHeader(http.StatusInternalServerError)
	}
}
