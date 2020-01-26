package main

import (
	"net/http"
	"os"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/cmd"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	go promMetrics()

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	log.Info("Starting up")

	cmd.Execute()
}

func promMetrics() {
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	log.Info("Serving /metrics endpoint.")
	// Start HTTP server
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
