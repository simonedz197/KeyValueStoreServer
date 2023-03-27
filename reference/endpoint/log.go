package endpoint

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func log(r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"time":   time.Now(),
		"source": r.RemoteAddr,
		"Method": r.Method,
		"URL":    r.URL.Path,
	}).Info("Request")
}

// Report any errors from an action.
func Report(action string, err error) {
	if err != nil {
		logrus.WithError(err).WithField("action", action).Error("Conversation error")
	}
}
