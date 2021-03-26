package main

import (
	"os"

	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
	"github.com/op/go-logging"
)

var log = logger.GetLogger(logging.DEBUG)

func main() {
	config, err := NewConfig()

	if err != nil {
		log.Errorf("Failed to load config")
		os.Exit(1)
	}

	log.Info("Parsed config successfully")

	panic(Init(*config))
}
