package repository

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/freecloudio/server/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	log "gopkg.in/clog.v1"
)

// ErrNeoNotInitialized is returned if a repository is initialized before the database connection
var ErrNeoNotInitialized = errors.New("db repository: neo4j repository must be initialized first")

var graphInfo *models.GraphInfo

type neoLabelConstraint struct {
	label       string
	model       interface{}
	uniqueProps []string
	exisProps   []string
}

// List of constraints filled in 'init' functions of each repository
var neoLabelConstraints []*neoLabelConstraint

var graphConnection neo4j.Driver

// InitGraphDatabaseConnection connects to the give neo4j server
func InitGraphDatabaseConnection(url, username, password string) error {
	driver, err := neo4j.NewDriver(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return err
	}
	graphConnection = driver

	for _, labelConstraint := range neoLabelConstraints {
		createConstraintForLabel(labelConstraint)
	}

	return nil
}

// CloseGraphDatabaseConnection closes the current connection to neo4j
func CloseGraphDatabaseConnection() (err error) {
	if err = graphConnection.Close(); err != nil {
		log.Error(0, "Failed to close neo4j connection: %v", err)
		return
	}

	return
}

// GetGraphInfo returns info about the currenlty connected neo4j database
func GetGraphInfo() (graphInfoRet *models.GraphInfo, err error) {
	if graphInfo != nil {
		return graphInfo, err
	}

	session, err := GetGraphSession()
	if err != nil {
		return
	}
	graphInfoRet = &models.GraphInfo{}

	res, err := session.Run("CALL dbms.queryJmx(\"org.neo4j:instance=kernel#0,name=Configuration\")", nil)
	if err != nil {
		log.Error(0, "Failed to run version/edition query: %v", err)
		return
	}
	summary, err := res.Summary()
	if err != nil {
		log.Error(0, "Failed to get summary or result of version/edition query: %v", err)
		return
	}
	graphInfoRet.Version = summary.Server().Version()

	rec, err := neo4j.Single(res, nil)
	if err != nil {
		log.Error(0, "Failed to get record from edition query: %v", err)
		return
	}
	configMapInt, ok := rec.Get("attributes")
	if !ok {
		err = errors.New("failed to get attributes from edition query")
		log.Error(0, err.Error())
		return
	}
	configMap := configMapInt.(map[string]interface{})
	editionMapInt, ok := configMap["unsupported.dbms.edition"]
	if !ok {
		err = errors.New("failed to get edition map from config map")
		log.Error(0, err.Error())
		return
	}
	editionMap := editionMapInt.(map[string]interface{})
	graphInfoRet.Edition, ok = editionMap["value"].(string)
	if !ok {
		err = errors.New("failed to get value map from edition map")
		log.Error(0, err.Error())
		return
	}

	graphInfo = graphInfoRet

	return
}

// GetGraphSession return a new neo4j session
func GetGraphSession() (neo4j.Session, error) {
	sess, err := graphConnection.Session(neo4j.AccessModeWrite)
	if err != nil {
		log.Error(0, "Failed to create neo4j session: %v", err)
	}
	return sess, err
}

func createConstraintForLabel(labelConstraint *neoLabelConstraint) error {
	modelValue := reflect.ValueOf(labelConstraint.model).Elem()
	modelType := modelValue.Type()

	for it := 0; it < modelType.NumField(); it++ {
		typeField := modelType.Field(it)
		dbNamePtr := getDBFieldName(typeField)
		if dbNamePtr == nil {
			continue
		}
		dbName := *dbNamePtr
		_, isUnique := typeField.Tag.Lookup("fc_neo_unique")

		labelConstraint.exisProps = append(labelConstraint.exisProps, dbName)
		if isUnique {
			labelConstraint.uniqueProps = append(labelConstraint.uniqueProps, dbName)
		}
	}

	session, err := GetGraphSession()
	if err != nil {
		return err
	}

	uniqueQuery := "CREATE CONSTRAINT ON (c:%s) ASSERT c.%s IS UNIQUE"
	for _, uniqueProp := range labelConstraint.uniqueProps {
		_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(fmt.Sprintf(uniqueQuery, labelConstraint.label, uniqueProp), nil)
		})
		if err != nil {
			log.Error(0, "Failed to create unique constraint on label %s with property %s: %v", labelConstraint.label, uniqueProp, err)
			continue
		}
	}

	info, err := GetGraphInfo()
	if err != nil || info.Edition == "community" {
		return err
	}

	exisQuery := "CREATE CONSTRAINT ON (c:%s) ASSERT exists(c.%s)"
	for _, exisProp := range labelConstraint.exisProps {
		_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(fmt.Sprintf(exisQuery, labelConstraint.label, exisProp), nil)
		})
		if err != nil {
			log.Error(0, "Failed to create exist constraint on label %s with property %s: %v", labelConstraint.label, exisProp, err)
			continue
		}
	}

	return nil
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
	if !ok {
		return nil, errors.New("value not found with key '" + key + "'")
	}
	valNode, ok := valInt.(neo4j.Node)
	if !ok {
		return nil, errors.New("value with key '" + key + "' could not be converted to 'neo4j.Node'")
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
