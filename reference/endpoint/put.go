package endpoint

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reference/store"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/gorilla/mux"
)

// Put will attempt to store a value under a key.
func Put(lru *store.LRU) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log(r)

		key := mux.Vars(r)["key"]
		user := session.Get(r)["user"]

		b, err := ioutil.ReadAll(r.Body)

		if err != nil || len(b) == 0 {
			Report("put", conversation.Respond(w, http.StatusBadRequest))
			return
		}

		defer func() { _ = r.Body.Close() }()

		err = lru.Put(key, string(b), user)

		switch {
		case errors.Is(err, store.ErrNotOwner):
			Report("put", conversation.Respond(w, http.StatusForbidden))
		case errors.Is(err, store.ErrNotFound):
			Report("put", conversation.Respond(w, http.StatusNotFound))
		default:
			Report("put", conversation.Respond(w, http.StatusOK))
		}
	}
}
