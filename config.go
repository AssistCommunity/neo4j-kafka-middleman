package main

import (
	"os"

	kafkaIntegration "github.com/AssistCommunity/neo4j-kafka-middleman/kafka"
	"github.com/AssistCommunity/neo4j-kafka-middleman/neo4jIntegration"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Neo4j neo4jIntegration.Neo4jConfig `yaml:"neo4j"`
	Kafka kafkaIntegration.KafkaConfig `yaml:"kafka"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	file, err := os.Open("./config.yaml")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
