package main

import (
	"net/http"
	"os"

	"github.com/omnikk/pz4.2/auth/internal/handler"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	entry := log.WithField("service", "auth")

	h := handler.New(entry)
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/verify", h.Verify)

	entry.Info("auth service starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		entry.WithField("error", err.Error()).Fatal("server failed")
	}
}
