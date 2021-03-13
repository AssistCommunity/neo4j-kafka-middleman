package main

import (
	"encoding/json"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func main() {

	// Kafka config params
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = "go-kafka-middleman"
	kafkaConfig.Consumer.Return.Errors = true

	kafkaBrokers := []string{"b-2.assistkafka.9y2zsl.c3.kafka.ap-south-1.amazonaws.com:9092", "b-1.assistkafka.9y2zsl.c3.kafka.ap-south-1.amazonaws.com:9092"}

	kafkaConsumer, err := sarama.NewConsumer(kafkaBrokers, kafkaConfig)

	if err != nil {
		panic(err)
	}

	neo4jConnString := "bolt://52.66.229.170:7687"

	driver, err := neo4j.NewDriver(neo4jConnString, neo4j.BasicAuth("neo4j", "neo4j", ""))

	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity()

	if err != nil {
		panic(err)
	}

	topics, _ := kafkaConsumer.Topics()

	messages, errors := consume(topics, kafkaConsumer)

	go func() {
		for {
			err = <-errors
		}
	}()

	neo4jErrors := make(chan error)

	go neo4jProcessMessage(messages, neo4jErrors)

}

func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	messages := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)
		// this only consumes partition no 1, you would probably want to consume all partitions
		for _, partition := range partitions {
			consumer, err := master.ConsumePartition(topic, partition, sarama.OffsetOldest)

			if err != nil {
				panic(err)
			}

			go consumePartition(consumer, messages, errors)
		}
	}

	return messages, errors
}

func consumePartition(consumer sarama.PartitionConsumer, messages chan *sarama.ConsumerMessage, errors chan *sarama.ConsumerError) {
	for {
		select {
		case consumeError := <-consumer.Errors():
			errors <- consumeError
		case message := <-consumer.Messages():
			messages <- message
		}
	}
}

func neo4jProcessMessage(messages chan *sarama.ConsumerMessage, errors chan error) {
	for {
		message := <-messages

		topic := message.Topic

		query, err := QueryFromTopic(topic)
		if err != nil {
			errors <- err
			continue
		}

		messageBody := message.Value

		params := make(map[string]interface{})

		json.Unmarshal(messageBody, params)

		err = Neo4jRunQuery(query, params)
		if err != nil {
			errors <- err
			continue
		}
	}
}
