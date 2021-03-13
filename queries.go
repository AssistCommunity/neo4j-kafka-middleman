package main

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	SyncContacts = "sync-contacts"
)

var topicQueryMap = map[string]string{
	"post-upvote": `
		CREATE (n:TestNode {key: "value"}) RETURN n
	`,
}

func QueryFromTopic(topic string) (string, error) {

	query, ok := topicQueryMap[topic]

	if !(ok) {
		err := fmt.Errorf("cannot find query corresponding to topic: %s", topic)
		return "", err
	}

	return query, nil
}

func Neo4jRunQuery(session *LockableNeo4jSession, query string, params map[string]interface{}) error {

	session.mu.Lock()
	_, err := session.session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(query, params)

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record(), nil
		}

		return nil, result.Err()
	})
	session.mu.Unlock()

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
