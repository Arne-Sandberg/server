package repository

import (
	"testing"

	"github.com/freecloudio/server/config"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func TestInitGraphDatabaseConnection(t *testing.T) {
	defer testCloseClearGraph()

	err := InitGraphDatabaseConnection(config.GetString("graph.url"), config.GetString("graph.user"), config.GetString("graph.password"))
	if err != nil {
		t.Fatalf("Failed to init graph database connection: %v", err)
	}

	if graphConnection == nil {
		t.Errorf("After successfull graph database initialization the connection variable is still null")
	}
}

func TestGetGraphSession(t *testing.T) {
	testConnectClearGraph()
	defer testCloseClearGraph()

	sess, err := getGraphSession()
	if err != nil {
		t.Fatalf("Failed to get graph session: %v", err)
	}
	sess.Close()
}

func TestCloseGraphDatabaseConnection(t *testing.T) {
	testConnectClearGraph()

	err := CloseGraphDatabaseConnection()
	if err != nil {
		t.Fatalf("Failed to close graph database connection: %v", err)
	}
}

func testClearGraph() {
	sess, _ := getGraphSession()
	defer sess.Close()
	sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (n) DETACH DELETE n", nil)
	})
}

func testConnectClearGraph() {
	InitGraphDatabaseConnection(config.GetString("graph.url"), config.GetString("graph.user"), config.GetString("graph.password"))
	testClearGraph()
}

func testCloseClearGraph() {
	testClearGraph()
	CloseGraphDatabaseConnection()
}
