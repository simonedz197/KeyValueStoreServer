// Package handlers handles serving lists.
package handlers

import (
	log "KeyValueStoreServer/server/loggers"
	store "KeyValueStoreServer/server/store"
	"encoding/json"
	"net/http"
	"strings"
)

const elementLimit = 2

// ServeList - returns a list of all keys and owners.
func ServeList(writer http.ResponseWriter, req *http.Request) {
	log.RequestChannel <- req

	// see if we have an additional key value in url
	// otherwise just return complete list
	path, _ := strings.CutPrefix(req.URL.Path, "/")
	path, _ = strings.CutSuffix(path, "/")
	elements := strings.Split(path, "/")

	if len(elements) > elementLimit {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	username := getUsername(req)

	if username == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte("Not Authorised"))

		return
	}

	var (
		data  []byte
		count int
	)

	if len(elements) == 1 {
		// get complete list
		data, count = getList(username)
	} else {
		// get for specific key

		data, count = getListForKey(elements[1], username)

		if count == 0 {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte("404 Key Not Found"))

			return
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if count > 0 {
		_, _ = writer.Write(data)
	} else {
		_, _ = writer.Write([]byte("[]"))
	}
}

func getList(username string) ([]byte, int) {
	// get complete list
	list := <-store.PublicAccess.List(username)
	jsonData, err := json.Marshal(list)

	if err == nil {
		return jsonData, len(list)
	}

	return nil, 0
}

func getListForKey(username, key string) ([]byte, int) {
	// get complete list
	list := <-store.PublicAccess.ListForKey(username, key)

	if len(list) > 0 {
		jsonData, err := json.Marshal(list[0])

		if err == nil {
			return jsonData, len(list)
		}
	}

	return nil, 0
}
