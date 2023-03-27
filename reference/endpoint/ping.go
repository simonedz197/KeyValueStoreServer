package endpoint

import (
	"net/http"

	"bitbucket.org/idomdavis/gohttp/conversation"
)

// Ping responds with pong.
func Ping(w http.ResponseWriter, r *http.Request) {
	log(r)
	Report("pong", conversation.Reply(w, "pong"))
}
