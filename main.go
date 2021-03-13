package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type LockableNeo4jSession struct {
	session neo4j.Session
	mu      sync.Mutex
}

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

	driver, err := neo4j.NewDriver(neo4jConnString, neo4j.BasicAuth("neo4j", "assistneo4jpassword", ""))

	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity()

	if err != nil {
		panic(err)
	}

	neo4jSession := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	safeNeo4jSession := &LockableNeo4jSession{
		session: neo4jSession,
	}
	topics, _ := kafkaConsumer.Topics()

	kafkaMessages, errors := consume(topics, kafkaConsumer)

	neo4jErrors := make(chan error)

	go neo4jProcessMessage(safeNeo4jSession, kafkaMessages, neo4jErrors)

	for {
		err := <-errors
		fmt.Printf("%s\n", err)
	}
}

func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	messages := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.HasPrefix(topic, "__") {
			continue
		}
		partitions, _ := master.Partitions(topic)
		// this only consumes partition no 1, you would probably want to consume all partitions
		for _, partition := range partitions {
			fmt.Println("Here")
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

func neo4jProcessMessage(session *LockableNeo4jSession, messages chan *sarama.ConsumerMessage, errors chan error) {
	for {
		message := <-messages

		fmt.Println(message)
	}
}
