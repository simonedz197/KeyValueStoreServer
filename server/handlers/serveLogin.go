// Package handlers serve login.
package handlers

import (
	"net/http"
)

// ServeLogin serve to login endpoint.
func ServeLogin(writer http.ResponseWriter, r *http.Request) {
	_, ok, tokenString := authenticateUserFromBasicAuth(r)
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte("Unauthorised"))

		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("Bearer " + tokenString))
}
