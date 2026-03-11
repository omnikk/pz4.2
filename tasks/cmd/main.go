package main

import (
	"net/http"
	"os"

	"github.com/omnikk/pz4.2/tasks/internal/handler"
	"github.com/omnikk/pz4.2/tasks/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "ts",
			logrus.FieldKeyMsg:  "msg",
		},
	})
	log.SetOutput(os.Stdout)
	entry := log.WithField("service", "tasks")

	metrics.Register()

	authAddr := "http://localhost:8081"

	h := handler.New(entry, authAddr)
	mux := http.NewServeMux()

	h.Register(mux)
	mux.Handle("/metrics", promhttp.Handler())

	entry.Info("tasks service starting on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		entry.WithField("error", err.Error()).Fatal("server failed")
	}
}
