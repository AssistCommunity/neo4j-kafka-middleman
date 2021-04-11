package kafkaIntegration

import (
	"context"

	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
	"github.com/Shopify/sarama"
)

var log = logger.GetLogger()

type KafkaConfig struct {
	GroupID string   `yaml:"group_id"`
	Brokers []string `yaml:"brokers"`
}

type Handler struct {
	messages chan *sarama.ConsumerMessage
	errors   chan error
}

func GetClient(config KafkaConfig) (sarama.ConsumerGroup, error) {
	kafkaConfig := sarama.NewConfig()

	kafkaConfig.Consumer.Return.Errors = true

	kafkaBrokers := config.Brokers
	kafkaGroup := config.GroupID

	client, err := sarama.NewConsumerGroup(kafkaBrokers, kafkaGroup, kafkaConfig)

	return client, err
}

func Consume(topics []string, client *sarama.ConsumerGroup) (chan *sarama.ConsumerMessage, chan error) {
	messages := make(chan *sarama.ConsumerMessage)
	errors := make(chan error)

	handler := &Handler{
		messages,
		errors,
	}

	ctx := context.Background()
	go func() {
		for {
			if err := (*client).Consume(ctx, topics, handler); err != nil {
				handler.errors <- err
				return
			}

			if err := ctx.Err(); err != nil {
				handler.errors <- err
				return
			}
		}
	}()

	return messages, errors
}

func (handler *Handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		handler.messages <- message
		log.Infof("Message claimed: value = %s timestamp = %v topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "read")
	}

	return nil
}
