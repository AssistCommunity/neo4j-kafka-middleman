package main

import (
	"os"

	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
)

var log = logger.GetLogger()

func main() {
	config, err := NewConfig()

	if err != nil {
		log.Errorf("Failed to load config")
		os.Exit(1)
	}

	print("Hello, world!")

	log.Debugf("config: %+v\n", config)

	panic(Init(*config))
}
