package kafka

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
)

type KafkaConfig struct {
	ClientID string `yaml:"client_id"`

	Brokers []string `yaml:"brokers"`
}

func GetConsumer(config KafkaConfig) (sarama.Consumer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = config.ClientID

	kafkaConfig.Consumer.Return.Errors = true

	kafkaBrokers := config.Brokers

	consumer, err := sarama.NewConsumer(kafkaBrokers, kafkaConfig)
	fmt.Printf("%+v\n", consumer)
	return consumer, err
}

func Consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
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
			consumer, err := master.ConsumePartition(topic, partition, sarama.OffsetNewest)

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
