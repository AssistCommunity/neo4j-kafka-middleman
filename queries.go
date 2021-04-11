package main

import (
	"fmt"
	"time"

	"github.com/AssistCommunity/neo4j-kafka-middleman/logger"
	"github.com/AssistCommunity/neo4j-kafka-middleman/neo4jIntegration"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	SyncContacts = "sync-contacts"
)

type TopicToQuery map[string]string

func (t *TopicToQuery) GetQuery(topic string) (string, error) {
	if query, ok := (*t)[topic]; ok {
		return query, nil
	}
	return "", fmt.Errorf("cannot find matching query for topic: %s", topic)
}

func (t *TopicToQuery) GetTopics() []string {
	keys := make([]string, 0, len(*t))
	for k := range *t {
		keys = append(keys, k)
	}

	return keys
}

func Neo4jRunQuery(session *neo4jIntegration.LockableNeo4jSession, query string, params map[string]interface{}) error {

	session.Mu.Lock()
	start := time.Now()
	_, err := session.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(query, params)

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	elapsed := time.Since(start)
	session.Mu.Unlock()

	if err != nil {
		fmt.Println(err)
		return err
	}

	var log = logger.GetLogger()
	log.Infof("Neo4j query executed in %s", elapsed)

	return nil
}
