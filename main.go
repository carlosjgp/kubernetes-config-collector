package main

import (
	"os"

	"github.com/carlosjgp/kubernetes-config-collector/pkg/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	cmd.Execute()
}
