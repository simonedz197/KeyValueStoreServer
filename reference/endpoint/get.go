package endpoint

import (
	"net/http"
	"reference/store"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"github.com/gorilla/mux"
)

// Get returns a value stored under a key. An error response is sent if the key
// doesn't exist.
func Get(lru *store.LRU) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log(r)

		key := mux.Vars(r)["key"]

		if v, err := lru.Get(key); err != nil {
			Report("get", conversation.Respond(w, http.StatusNotFound))
		} else {
			Report("get", conversation.Reply(w, v))
		}
	}
}
