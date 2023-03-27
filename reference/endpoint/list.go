package endpoint

import (
	"net/http"
	"reference/store"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"github.com/gorilla/mux"
)

// List will return details of keys in the store. If the store is an LRU store
// then a detailed list will be returned.
func List(lru *store.LRU) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log(r)

		key := mux.Vars(r)["key"]

		if key == "" {
			Report("list all", conversation.Reply(w, lru.ListAll()))
		} else if entry, err := lru.List(key); err != nil {
			Report("list", conversation.Respond(w, http.StatusNotFound))
		} else {
			Report("list", conversation.Reply(w, entry))
		}
	}
}
