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

var topicQueryMap = map[string]string{
	"sync-contacts": `
		MATCH (CU:User {email: $email})
		WITH CU
		UNWIND $contacts AS contact
		MERGE (u:User {phone_number: contact.phone_number}) WHERE contact.phone_number <> u.phone_number
		ON CREATE SET u.app=false
		MERGE (CU)-[r:HAS_CONTACT]->(u)
		SET r.contact_name = contact.name, r.synced_on = CASE WHEN r.date IS NULL THEN datetime() ELSE r.date END
		MERGE (CU)-[s:AFFINITY_EDGE]->(u)
		ON CREATE SET s.total_score = 0, s.follows_score = 0, s.year_score = 0, s.branch_score = 0, s.hostel_score = 0
		SET s.contact_score = 1
		RETURN CU 
	`,
	"user-init": `
		CREATE (CU:User {email: $email, phone_number: $phone_number, on_app: true}) 
		WITH CU
		MATCH (u:User {on_app: true}) WHERE u.email <> CU.email
		WITH CU, u
		CREATE (CU)-[r1:AFFINITY_EDGE]->(u)
		WITH CU, u, r1
		CREATE (CU)<-[r2:AFFINITY_EDGE]-(u)
		SET r1.total_score = 0, r1.contact_score = 0, r1.follows_score = 0, r1.year_score = 0, r1.branch_score = 0, r1.hostel_score = 0
		SET r2.total_score = 0, r2.contact_score = 0, r2.follows_score = 0, r2.year_score = 0, r2.branch_score = 0, r2.hostel_score = 0
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
		SET CU.branches = $branches, CU.hostel = $hostel, CU.year = $year
		WITH CU
		MATCH (u:User {on_app: true})
		MATCH (CU)-[r1:AFFINITY_EDGE]->(u)
		MATCH (u)-[r2:AFFINITY_EDGE]->(CU) 
		WITH CU, u, r1, r2
		SET r1.year_score = CASE WHEN CU.year = u.year THEN 1 ELSE 0 END,
			r1.hostel_score = CASE WHEN CU.hostel = u.hostel THEN 1 ELSE 0 END,
			r2.year_score = CASE WHEN CU.year = u.year THEN 1 ELSE 0 END,
			r2.hostel_score = CASE WHEN CU.hostel = u.hostel THEN 1 ELSE 0 END

		WITH CU, r1, r2, [n IN CU.branches WHERE n IN u.branches] as commonBranches
		SET r1.branch_score = size(commonBranches)
		SET r2.branch_score = r1.branch_score
		RETURN CU
	`,

	"calculate-affinities": `
		MATCH (u1:User {email: $email})-[r:AFFINITY_EDGE]->(u2:User)
		SET r.total_score = 0.2*r.year_score + 0.3*r.branch_score + 0.3*r.hostel_score + 0.4*r.contact_score + 0.5*r.follows_score
		RETURN r
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
