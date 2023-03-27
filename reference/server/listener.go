package server

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Listen on the given port for HTTP requests.
func Listen(port int, handler http.Handler) error {
	logrus.WithField("port", port).Info("Listening for HTTP connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
