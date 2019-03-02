package repository

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	log "gopkg.in/clog.v1"
)

var (
	// ErrRecordNotFound indicated that no results were found for a query
	ErrRecordNotFound = errors.New("Neo4j record not found")
)

var graphConnection neo4j.Driver

// InitGraphDatabaseConnection connects to the give neo4j server
func InitGraphDatabaseConnection(url, username, password string) error {
	driver, err := neo4j.NewDriver(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return err
	}
	graphConnection = driver

	return nil
}

// CloseGraphDatabaseConnection closes the current connection to neo4j
func CloseGraphDatabaseConnection() error {
	if err := graphConnection.Close(); err != nil {
		log.Error(0, "Failed to close neo4j connection: %v", err)
	}

	return nil
}

// getGraphSession return a new neo4j session
func getGraphSession() (neo4j.Session, error) {
	sess, err := graphConnection.Session(neo4j.AccessModeWrite)
	if err != nil {
		log.Error(0, "Failed to create neo4j session: %v", err)
	}
	return sess, err
}
