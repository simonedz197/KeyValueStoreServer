package server

import (
	"net/http"

	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/gorilla/mux"
)

// Router builds the http router providing secure and insecure sub routers.
func Router(routes Routes, signatory session.Signatory) http.Handler {
	router := mux.NewRouter()
	secure := router.NewRoute().Subrouter()
	secure.Use(Authenticator(signatory))

	for method, handlers := range routes.Insecure {
		for p, h := range handlers {
			router.HandleFunc(p, h).Methods(method)
		}
	}

	for method, handlers := range routes.Secure {
		for p, h := range handlers {
			secure.HandleFunc(p, h).Methods(method)
		}
	}

	return router
}
