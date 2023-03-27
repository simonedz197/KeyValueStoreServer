package endpoint

import (
	"net/http"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/session"
)

// NEVER do this. Like ever. Argon2 or Bcrypt the passwords.
var users = map[string]string{
	"user_a": "passwordA",
	"user_b": "passwordB",
	"user_c": "passwordC",
	"admin":  "Password1",
}

// Login responds with a bearer token, or a 401.
func Login(signatory session.Signatory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log(r)

		username, passphrase, _ := r.BasicAuth()

		if users[username] != passphrase {
			Report("login", conversation.Respond(w, http.StatusUnauthorized))
			return
		}

		if err := signatory.Sign(w, session.Context{"user": username}); err != nil {
			Report("login", conversation.Respond(w, http.StatusInternalServerError))
			return
		}

		Report("login", conversation.Respond(w, http.StatusOK))
	}
}
