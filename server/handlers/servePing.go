// Package handlers handle ping.
package handlers

import (
	log "KeyValueStoreServer/server/loggers"
	"fmt"
	"net/http"
)

// ServePing return ok pong if called with a get otherwise returns method not allowed.
func ServePing(writer http.ResponseWriter, r *http.Request) {
	log.RequestChannel <- r

	if r.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writer.Header().Set("Content-Type", "text/plain charset=utf-8")

	_, err := writer.Write([]byte("pong"))
	if err != nil {
		warnMessage := fmt.Sprintf("Could not write to response %v", err)
		log.WarnChannel <- warnMessage

		writer.WriteHeader(http.StatusInternalServerError)
	}
}
