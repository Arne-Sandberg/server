package manager

import (
	"github.com/freecloudio/server/config"
	"github.com/freecloudio/server/repository"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func testClearGraph() {
	sess, _ := repository.GetGraphSession()
	defer sess.Close()
	sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (n) DETACH DELETE n", nil)
	})
}

func testConnectClearGraph() {
	repository.InitGraphDatabaseConnection(config.GetString("graph.url"), config.GetString("graph.user"), config.GetString("graph.password"))
	testClearGraph()
}

func testCloseClearGraph() {
	testClearGraph()
	repository.CloseGraphDatabaseConnection()
}
