package server

import (
	"net/http"
	"reference/endpoint"
	"reference/store"

	"bitbucket.org/idomdavis/gohttp/session"
)

// Handlers defines a set of Gorilla URL paths and their Handler.
type Handlers map[string]http.HandlerFunc

// Routes allows the API to be defined.
type Routes struct {
	Secure   map[string]Handlers
	Insecure map[string]Handlers
}

// API for the store.
func API(lru *store.LRU, signatory session.Signatory) Routes {
	return Routes{
		Secure: map[string]Handlers{
			http.MethodGet: map[string]http.HandlerFunc{
				"/store/{key}": endpoint.Get(lru),
				"/list":        endpoint.List(lru),
				"/list/{key}":  endpoint.List(lru),
				"/shutdown":    endpoint.Shutdown,
			},
			http.MethodPut: map[string]http.HandlerFunc{
				"/store/{key}": endpoint.Put(lru),
			},
			http.MethodDelete: map[string]http.HandlerFunc{
				"/store/{key}": endpoint.Delete(lru),
			},
		},
		Insecure: map[string]Handlers{
			http.MethodGet: map[string]http.HandlerFunc{
				"/ping":   endpoint.Ping,
				"/coffee": endpoint.Coffee,
				"/login":  endpoint.Login(signatory),
			},
		},
	}
}
