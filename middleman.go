package main

import (
	"encoding/json"
	"fmt"

	"github.com/AssistCommunity/neo4j-kafka-middleman/kafkaIntegration"
	"github.com/AssistCommunity/neo4j-kafka-middleman/neo4jIntegration"
	"github.com/Shopify/sarama"
)

func Init(config Config) error {

	// Init neo4j driver
	neo4jDriver, err := neo4jIntegration.GetDriver(config.Neo4j)

	if err != nil {
		return err
	}
	safeNeo4jSession := neo4jIntegration.GetLockableSession(neo4jDriver)

	// Init kafka driver
	kafkaConsumer, err := kafkaIntegration.GetConsumer(config.Kafka)

	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	topics, err := kafkaConsumer.Topics()

	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	kafkaMessages, errors := kafkaIntegration.Consume(topics, kafkaConsumer)

	go neo4jProcessMessage(config.TopicToQuery, &safeNeo4jSession, kafkaMessages)

	for {
		err := <-errors
		log.Errorf("%s\n", err)
	}
}

func neo4jProcessMessage(t TopicToQuery, session *neo4jIntegration.LockableNeo4jSession, messages chan *sarama.ConsumerMessage) {
	for {
		message := <-messages

		topic := message.Topic

		query, err := t.GetQuery(topic)

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
