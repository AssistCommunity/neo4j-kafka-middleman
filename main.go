package main

import (
	"flag"
	"os"

	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
)

var log = logger.GetLogger()

func main() {
	log.Info("Starting Neo4j <> Kafka middleman...")

	confPath := flag.String("config", "/conf/config.yaml", "The path to the config file to be used")

	flag.Parse()

	log.Infof("Using config file: ", *confPath)

	config, err := NewConfig(*confPath)

	if err != nil {
		log.Errorf("Failed to load config: %s", err)
		os.Exit(1)
	}

	log.Info("Parsed config successfully")
	log.Debugf("Neo4j config: %+v", config.Neo4j)
	log.Debugf("Kafka config: %+v", config.Kafka)

	panic(Init(*config))
}
