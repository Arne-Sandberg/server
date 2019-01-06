package repository

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "gopkg.in/clog.v1"
)

var ErrGormNotInitialized = errors.New("db repository: gorm repository must be initialized first")

// databaseConnection is shared between most repositories (session, user, ...)
var (
	databaseConnection *gorm.DB
	databaseModels     []interface{} // Contains pointers to all models that should be automigrated by gorm initialization
)

func InitDatabaseConnection(databaseType, user, password, host string, port int, name string) error {
	var args string

	switch databaseType {
	case "mysql":
		args = fmt.Sprintf("%s:%s@%s:%d/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, name)
		break
	case "postgres":
		args = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s", host, port, user, name, password)
		break
	case "mssql":
		args = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", user, password, host, port, name)
		break
	case "sqlite3":
		fallthrough
	default:
		databaseType = "sqlite3"
		args = name
	}

	db, err := gorm.Open(databaseType, args)
	if err != nil {
		log.Error(0, "Could not open datbase: %v", err)
		return err
	}
	log.Info("Initialized database connection")

	err = db.AutoMigrate(databaseModels...).Error
	if err != nil {
		log.Error(0, "Failed to auto migrate db structs: %v", err)
		return err
	}
	databaseConnection = db

	return nil
}

func CloseDatabaseConnection() {
	if err := databaseConnection.Close(); err != nil {
		log.Fatal(0, "Error shutting down gorm: %v", err)
		return
	}

	databaseConnection = nil
}

func IsRecordNotFoundError(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}
