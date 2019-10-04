package main

import (
	"net/http"

	"TraefikAccessControl/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	srv := server.NewServer()

	// Start
	log.Info("Listening on :4181")
	log.Info(http.ListenAndServe(":4181", srv.Router))
}
