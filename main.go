package main

import (
	"net/http"
	"os"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/cmd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9090", nil)

	cmd.Execute()
}
