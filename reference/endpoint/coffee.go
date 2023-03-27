package endpoint

import (
	"net/http"

	"bitbucket.org/idomdavis/gohttp/conversation"
)

// Coffee responds with I'm a teapot.
func Coffee(w http.ResponseWriter, r *http.Request) {
	log(r)
	Report("coffee", conversation.Respond(w, http.StatusTeapot))
}
