package repository

import (
	"reflect"
	"testing"

	"github.com/freecloudio/server/config"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func TestInitGraphDatabaseConnection(t *testing.T) {
	defer testCloseClearGraph()

	err := InitGraphDatabaseConnection(config.GetString("graph_url"), config.GetString("graph_user"), config.GetString("graph_password"))
	if err != nil {
		t.Fatalf("Failed to init graph database connection: %v", err)
	}

	if graphConnection == nil {
		t.Errorf("After successfull graph database initialization the connection variable is still null")
	}
}

func TestGetGraphInfo(t *testing.T) {
	testConnectClearGraph()
	defer testCloseClearGraph()

	info, err := GetGraphInfo()
	if err != nil {
		t.Fatalf("Failed to get graph session: %v", err)
	}
	if info.Version == "" || info.Edition == "" {
		t.Errorf("Version or edition of graph info empty: %v", info)
	}
}

func TestGetGraphSession(t *testing.T) {
	testConnectClearGraph()
	defer testCloseClearGraph()

	sess, err := GetGraphSession()
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

type testStruct struct {
	First  string `fc_neo:"other_first"`
	Second string `fc_neo:"-"`
	Third  string `json:"other_third"`
	Fourth string `json:"-"`
	Fifth  string
}

var testMap = map[string]interface{}{
	"other_first": "first_value",
	"other_third": "third_value",
	"Fifth":       "fifth_value",
}

func TestModelToMap(t *testing.T) {
	val := &testStruct{
		First:  "first_value",
		Second: "second_value",
		Third:  "third_value",
		Fourth: "fourth_value",
		Fifth:  "fifth_value",
	}

	resMap := modelToMap(val)

	if !reflect.DeepEqual(resMap, testMap) {
		t.Errorf("Result map and expected map not deeply equal: %v != %v", resMap, testMap)
	}
}

type testNode struct{}

func (node *testNode) Id() int64                     { return 0 }
func (node *testNode) Labels() []string              { return nil }
func (node *testNode) Props() map[string]interface{} { return testMap }

type testRecord struct{}

func (rec *testRecord) Keys() []string                   { return nil }
func (rec *testRecord) Values() []interface{}            { return nil }
func (rec *testRecord) GetByIndex(index int) interface{} { return nil }
func (rec *testRecord) Get(key string) (interface{}, bool) {
	switch key {
	case "t":
		return &testNode{}, true
	default:
		return nil, false
	}
}

func TestRecordToModel(t *testing.T) {
	expVal := &testStruct{
		First: "first_value",
		Third: "third_value",
		Fifth: "fifth_value",
	}

	resVal, err := recordToModel(&testRecord{}, "t", &testStruct{})
	if err != nil {
		t.Fatalf("Failed to convert record to model: %v", err)
	}

	if !reflect.DeepEqual(resVal, expVal) {
		t.Errorf("Result model value and expected model value not deeply equal: %v != %v", resVal, expVal)
	}
}

func testClearGraph() {
	sess, _ := GetGraphSession()
	defer sess.Close()
	sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (n) DETACH DELETE n", nil)
	})
}

func testConnectClearGraph() {
	InitGraphDatabaseConnection(config.GetString("graph_url"), config.GetString("graph_user"), config.GetString("graph_password"))
	testClearGraph()
}

func testCloseClearGraph() {
	testClearGraph()
	CloseGraphDatabaseConnection()
}
