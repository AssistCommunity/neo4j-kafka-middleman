package main

import (
	"encoding/json"
	"fmt"

	kafkaIntegration "github.com/AssistCommunity/neo4j-kafka-middleman/kafka"
	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
	"github.com/AssistCommunity/neo4j-kafka-middleman/neo4jIntegration"
	"github.com/Shopify/sarama"
)

var log = logger.GetLogger()

func main() {
	config, _ := NewConfig()
	log.Debugf("config: %+v\n", config)

	neo4jDriver, _ := neo4jIntegration.GetDriver(config.Neo4j)
	safeNeo4jSession := neo4jIntegration.GetLockableSession(neo4jDriver)

	kafkaConsumer, err := kafkaIntegration.GetConsumer(config.Kafka)

	if err != nil {
		log.Error(err)
	}

	topics, err := kafkaConsumer.Topics()

	if err != nil {
		log.Error(err)
	}

	kafkaMessages, errors := kafkaIntegration.Consume(topics, kafkaConsumer)

	go neo4jProcessMessage(&safeNeo4jSession, kafkaMessages)

	for {
		err := <-errors
		log.Errorf("%s\n", err)
	}
}

func neo4jProcessMessage(session *neo4jIntegration.LockableNeo4jSession, messages chan *sarama.ConsumerMessage) {
	for {
		message := <-messages

		topic := message.Topic

		query, err := QueryFromTopic(topic)

		if err != nil {
			fmt.Println(err)
			continue // replace with continue
		}

		var params map[string]interface{}

		err = json.Unmarshal(message.Value, &params)

		if err != nil {
			fmt.Println(err)
			continue // replace with continue
		}

		log.Infof("New Event on topic %s", topic)
		log.Debugf("Query: %s", query)
		Neo4jRunQuery(session, query, params)
	}
}
