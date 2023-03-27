package endpoint

import (
	"errors"
	"net/http"
	"reference/store"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/gorilla/mux"
)

// Delete will attempt to remove a value from key.
func Delete(lru *store.LRU) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log(r)

		key := mux.Vars(r)["key"]
		user := session.Get(r)["user"]
		err := lru.Delete(key, user)

		switch {
		case errors.Is(err, store.ErrNotOwner):
			Report("delete", conversation.Respond(w, http.StatusForbidden))
		case errors.Is(err, store.ErrNotFound):
			Report("delete", conversation.Respond(w, http.StatusNotFound))
		default:
			Report("delete", conversation.Respond(w, http.StatusOK))
		}
	}
}
