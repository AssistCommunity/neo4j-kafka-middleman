package main

import (
	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
)

var log = logger.GetLogger()

func main() {
	config, _ := NewConfig()

	log.Debugf("config: %+v\n", config)

	panic(Init(*config))
}
