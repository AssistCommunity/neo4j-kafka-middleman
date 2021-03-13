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
		CREATE (n:TestNode {key: value}) RETURN n
	`,

	"sync-contacts": `
		MATCH (CU:User {email: $email})
		WITH CU
		UNWIND $contacts as contact
		MERGE (u:User {phone_number: contact.phone_number}) WHERE u.phone_number <> CU.phone_number 
		ON CREATE 
			SET u.on_app = false
		MERGE (CU)-[r:HAS_CONTACT]->(u)
		SET r.contact_name = contact.name,
			r.synced_on = CASE WHEN r.date IS NULL THEN datetime() ELSE r.date END
		MERGE (CU)-[s:AFFINITY_EDGE]->(u)
		ON CREATE 
			SET s.total_score = 0
		SET s.has_contact_score = 1
		RETURN CU, r, u
	`,
	"user-init": `
		CREATE (CU:User {email: $email, phone_number: $phone_number, on_app: true}) 
		WITH CU
		MATCH (u:User {on_app: true}) WHERE u.email <> CU.email
		WITH CU, u
		CREATE (CU)-[r1:AFFINITY_EDGE]->(u)
		WITH CU, u, r1
		CREATE (CU)<-[r2:AFFINITY_EDGE]-(u)
		SET r1.total_score = 0, r2.total_score = 0
		RETURN CU 
	`,
	"follow-user": `
		MATCH (CU:User {email: $self_email}), (u:User {email: $target_email})
		MERGE (CU)-[r:FOLLOWS]->(u)
		MERGE (CU)-[s:AFFINITY_EDGE]->(u)
		SET s.follows_score = 1
		SET s.unfollows_score = 0
		
		RETURN CU, u
	`,
	"unfollow-user": `
		MATCH (CU:User {email: $self_email})-[r:FOLLOWS]->(u:User {email: $target_email})
		DELETE r
		MATCH (CU)-[s:AFFINITY_EDGE]->(u)
		SET s.follows_score = 0
		SET s.unfollows_score = 1

		RETURN CU, u
	`,
	"complete-profile": `

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
