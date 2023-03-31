// Package handlers common functions.
package handlers

import (
	log "KeyValueStoreServer/server/loggers"
	store "KeyValueStoreServer/server/store"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// Admin = admin.
const Admin = "admin"

type claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey = []byte("ebd4eca7-a114-478b-a12d-617d3a9d91e0")

func getUsername(r *http.Request) string {
	username, ok := authenticateUserFromToken(r)
	if !ok {
		username = ""
	}
	//	authHeader := r.Header.Get("Authorization")

	return username
}

// returns username and true if authenticated.
func authenticateUserFromToken(r *http.Request) (string, bool) {
	// first try authentication from bearer token. if that doesn't work
	// try basic auth. if that works create new token for future authentication
	// if doesn't work set unathorized
	rawTokenString := r.Header.Get("Authorization")
	tokenString, _ := strings.CutPrefix(rawTokenString, "Bearer ")

	claims := &claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) { return jwtKey, nil })

	if err != nil {
		log.WarnChannel <- fmt.Sprintf("Error processing JWT token %v", err)
		return "", false
	}

	return claims.Username, true
}

func authenticateUserFromBasicAuth(r *http.Request) (string, bool, string) {
	username, password, ok := r.BasicAuth()
	if !ok || username == "" || password == "" {
		log.ErrorChannel <- fmt.Sprintf("Unauthorised access attempt for user %s", username)
		return "", false, ""
	}

	valid := store.ValidateLogin(username, password)

	if !valid {
		log.ErrorChannel <- fmt.Sprintf("Unauthorised access attempt for user %s", username)
		return "", false, ""
	}

	// we are here so authorized.
	// build the jwt token
	expirationTime := time.Now().Add(time.Minute * 5)
	claims := &claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "UserJWTService",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", false, ""
	}

	return username, true, tokenString
}
