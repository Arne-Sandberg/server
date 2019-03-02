package repository

import (
	"errors"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	log "gopkg.in/clog.v1"
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

// Convert given struct to a map with the 'fc_neo' / 'json' / field name as key and the field value as value
func modelToMap(model interface{}) map[string]interface{} {
	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()
	modelMap := make(map[string]interface{})

	for it := 0; it < modelValue.NumField(); it++ {
		valField := modelValue.Field(it)
		typeField := modelType.Field(it)

		dbName := getDBFieldName(typeField)
		if dbName == nil {
			continue
		}
		modelMap[*dbName] = valField.Interface()
	}

	return modelMap
}

func recordToModel(record neo4j.Record, key string, model interface{}) (interface{}, error) {
	valInt, ok := record.Get(key)
	if ok == false {
		return nil, errors.New("Value not found with key '" + key + "'")
	}
	valNode, ok := valInt.(neo4j.Node)
	if ok == false {
		return nil, errors.New("Value with key '" + key + "' could not be converted to 'neo4j.Node'")
	}
	valProps := valNode.Props()

	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()

	for it := 0; it < modelValue.NumField(); it++ {
		valField := modelValue.Field(it)
		typeField := modelType.Field(it)

		dbNamePtr := getDBFieldName(typeField)
		if dbNamePtr == nil || !valField.CanSet() {
			continue
		}
		dbName := *dbNamePtr

		propInt, ok := valProps[dbName]
		if !ok {
			continue
		}
		propVal := reflect.ValueOf(propInt)
		valField.Set(propVal)
	}

	return model, nil
}

// Returns db field name based on tags of a struct field
// Returns nil if the field should not be stored in the db
// Uses own 'fc_neo' field tag but falls back to 'json' tags as these are automatically set from swagger
func getDBFieldName(typeField reflect.StructField) *string {
	var fieldTag string
	if fcNeoFieldTag := typeField.Tag.Get("fc_neo"); fcNeoFieldTag != "" {
		fieldTag = fcNeoFieldTag
	} else {
		fieldTag = strings.Split(typeField.Tag.Get("json"), ",")[0]
	}

	if fieldTag == "-" {
		return nil
	} else if fieldTag != "" {
		return &fieldTag
	} else {
		return &(typeField.Name)
	}
}
