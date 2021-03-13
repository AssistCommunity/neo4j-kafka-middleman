package main

import "fmt"

const (
	SyncContacts = "sync-contacts"
)

var topicQueryMap = map[string]string{
	"sync-contacts": `
		"MATCH (n) RETURN n"
	`,
	"upvote-post": `
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

func Neo4jRunQuery(query string, params map[string]interface{}) error {

	return nil
}
