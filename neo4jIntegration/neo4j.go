package neo4jIntegration

import (
	"sync"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type LockableNeo4jSession struct {
	Session neo4j.Session
	Mu      sync.Mutex
}

func GetDriver(config Neo4jConfig) (neo4j.Driver, error) {
	driver, err := neo4j.NewDriver(config.URI, neo4j.BasicAuth(config.Username, config.Password, ""))

	if err != nil {
		return nil, err
	}

	err = driver.VerifyConnectivity()

	if err != nil {
		return nil, err
	}

	return driver, err
}

func GetLockableSession(driver neo4j.Driver) LockableNeo4jSession {
	neo4jSession := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	return LockableNeo4jSession{
		Session: neo4jSession,
	}
}
