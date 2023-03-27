package endpoint

import (
	"net/http"
	"os"
	"time"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/sirupsen/logrus"
)

// Shutdown will exit the program.
func Shutdown(w http.ResponseWriter, r *http.Request) {
	log(r)

	if user := session.Get(r)["user"]; user == "admin" {
		Report("shutdown", conversation.Respond(w, http.StatusOK))
		logrus.WithField("user", user).Info("Shutdown requested")

		go func() {
			time.Sleep(time.Millisecond)
			os.Exit(0)
		}()
	} else {
		Report("shutdown", conversation.Respond(w, http.StatusForbidden))
	}
}
