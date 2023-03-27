package server

import (
	"net/http"
	"reference/endpoint"
	"strings"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/handler"
	"bitbucket.org/idomdavis/gohttp/middleware"
	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/gorilla/mux"
)

// Authenticator returns middleware that will check perform JWT authentication
// if a bearer is set, or simply set the user to the Authorisation header if no
// bearer is set. No auth header is an automatic 404.
func Authenticator(signatory session.Signatory) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		jwt := middleware.Authenticator{
			Signatory: signatory,
			Deny:      handler.Unauthorised{}.Handle,
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get(session.AuthHeader)
			switch {
			case auth == "":
				endpoint.Report("auth", conversation.Respond(w, http.StatusNotFound))
			case strings.HasPrefix(auth, "Bearer"):
				jwt.Authenticate(next).ServeHTTP(w, r)
			default:
				session.Store(r, session.Context{"user": auth})
				next.ServeHTTP(w, r)
			}
		})
	}
}
