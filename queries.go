package main

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	SyncContacts = "sync-contacts"
)

var topicQueryMap = map[string]string{
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
		MATCH (CU:User {email: $email})
		SET CU.branch1 = $branch1, CU.branch2 = $branch2, CU.hostel = $hostel, CU.year = $year
		MATCH (u:User {on_app: true})
		MATCH (CU)-[r1:AFFINITY_EDGE]->(u)
		MATCH (u)-[r2:AFFINITY_EDGE]->(CU) 
		WITH CU, u
		SET r1.year_score = CASE WHEN CU.year = u.year THEN 1 ELSE 0 END,
			r1.hostel_score = CASE WHEN CU.hostel = u.hostel THEN 1 ELSE 0 END,
			r2.year_score = CASE WHEN CU.year = u.year THEN 1 ELSE 0 END,
			r2.hostel_score = CASE WHEN CU.hostel = u.hostel THEN 1 ELSE 0 END,
		CALL apoc.do.case([
			CU.branch1 = u.branch2 XOR CU.branch2 = u.branch1 OR CU.branch1 = u.branch1 XOR CU.branch2 = u.branch1,
			'SET r1.branch_score = 1, r2.branch_score = 1',
			CU.branch1 = u.branch2 AND CU.branch2 = u.branch1 OR CU.branch1 = u.branch1 AND CU.branch2 = u.branch1,
			'SET r1.branch_score = 2, r2.branch_score = 2',			
		],
		'SET r1.branch_score = 0, r2.branch_score = 0', {CU: CU, u: u})
		YIELD value
		RETURN CU
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
