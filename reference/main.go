package main

import (
	"flag"
	"reference/server"
	"reference/store"
	"time"

	"bitbucket.org/idomdavis/gohttp/session"
	"github.com/sirupsen/logrus"
)

const (
	secret = "don't set these like this in prod, they need to get passed in"
	ttl    = time.Hour * 12
)

func main() {
	var (
		override    bool
		port, depth int
	)

	flag.IntVar(&port, "port", 8000, "port to listen on")
	flag.IntVar(&depth, "depth", 0, "LRU depth (zero for unbounded)")
	flag.BoolVar(&override, "override", false, "Enable admin override")
	flag.Parse()

	lru := store.NewLRU(depth)
	lru.Override = override

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		FullTimestamp:          true,
		PadLevelText:           true,
	})

	signatory := session.JWTSignatory{
		Secret: secret,
		TTL:    ttl,
	}

	router := server.Router(server.API(lru, signatory), signatory)

	if err := server.Listen(port, router); err != nil {
		logrus.WithError(err).Error("Server exited")
	} else {
		logrus.Info("Done")
	}
}
