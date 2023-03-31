// Package handlers server key based endpoints
package handlers

import (
	log "KeyValueStoreServer/server/loggers"
	store "KeyValueStoreServer/server/store"
	"errors"
	"io"
	"net/http"
	"strings"
)

// BaseURLPath /store.
const BaseURLPath = "/store"

// StoreKeyValueExistsForOwner 1.
const StoreKeyValueExistsForOwner = 1

// StoreKeyValueExistsButNotForOwner 2.
const StoreKeyValueExistsButNotForOwner = 2

// StoreKeyValueNotFound 3.
const StoreKeyValueNotFound = 3

// ServeKey marshalls the request to the appropriate handler based on method.
func ServeKey(writer http.ResponseWriter, req *http.Request) {
	log.RequestChannel <- req
	// we expect a key value on the url
	// get it here and pass it in to our worker functions
	// or just return a 400 bad request.
	key := GetKeyValue(BaseURLPath+"/", req.URL.Path)
	if key == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// if username not speficied in bearer token mak enot authorised
	username := getUsername(req)

	if username == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte("Not Authorised"))

		return
	}

	switch req.Method {
	case http.MethodGet:
		serveGet(writer, key, username)

	case http.MethodPut:
		value, err := io.ReadAll(req.Body)

		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		servePut(writer, string(value), key, username)

	case http.MethodDelete:
		serveDelete(writer, key, username)
	}
}

// GetKeyValue - takes a url path, removes the baseurl and creates an
// array using split
// if there is only 1 value in array returns it otherwise returns ""
// e.g. /store/mykey returns mykey
// e.g. /store/mykey/some/extra/stuff returns ”.
func GetKeyValue(prefix, urlPath string) string {
	path := strings.Replace(urlPath, prefix, "", 1)
	if path == "" {
		return path
	}

	elements := strings.Split(path, "/")
	if len(elements) != 1 {
		return ""
	}

	return elements[0]
}

// servePut - allowed to create or update a values in the store
// for the given key
// if updating the store entry must have been created by the username in basicauth
// otherwise return forbidden.
func servePut(writer http.ResponseWriter, value string, key string, owner string) {
	// Create upsert request message
	response := <-store.PublicAccess.Upsert(key, owner, value)

	if response != nil {
		if val, ok := response.(error); ok {
			switch {
			case errors.Is(val, store.ErrForbidden):
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte("Forbidden"))
			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}

			return
		}
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("OK"))
}

// serveGet - retreives a value for the given key
// all entries are accessible regardless of who created them
// if entry for key does not exist returns 404.
func serveGet(writer http.ResponseWriter, key string, owner string) {
	fetchResponse := <-store.PublicAccess.Fetch(key)

	dataval, ok := fetchResponse.(store.DataValue)
	if !ok || dataval.Owner != owner {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("404 key not found"))

		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(dataval.Value))
}

// serveDelete - deletes an entry for a given key
// only allowed if entry created by username
// if entry does not exist return 404
// if entry exists but belongs to a different username return 403 forbidden.
func serveDelete(writer http.ResponseWriter, key string, owner string) {
	// check key status
	response := <-store.PublicAccess.Delete(key, owner)
	if response != nil {
		if val, ok := response.(error); ok {
			switch {
			case errors.Is(val, store.ErrForbidden):
				writer.WriteHeader(http.StatusForbidden)
				_, _ = writer.Write([]byte("Forbidden"))
			case errors.Is(val, store.ErrNotFound):
				writer.WriteHeader(http.StatusNotFound)
				_, _ = writer.Write([]byte("404 key not found"))

			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}

		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok"))
}
