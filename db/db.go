package db

import (
	"github.com/cloudfoundry-community/aws-broker/common"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"errors"
	"fmt"
	"github.com/cloudfoundry-community/aws-broker/base"
	"github.com/cloudfoundry-community/aws-broker/services/rds"
	"log"
)

// DBinit is a generic helper function that will try to connect to a database with the config in the input.
// Supported DB types:
// * postgres
// * sqlite3
func DBInit(dbConfig *common.DBConfig) (*gorm.DB, error) {
	var DB gorm.DB
	var err error
	switch dbConfig.DbType {
	case "postgres":
		conn := "dbname=%s user=%s password=%s host=%s sslmode=%s port=%d"
		conn = fmt.Sprintf(conn,
			dbConfig.DbName,
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Url,
			dbConfig.Sslmode,
			dbConfig.Port)
		DB, err = gorm.Open("postgres", conn)
	case "sqlite3":
		DB, err = gorm.Open("sqlite3", dbConfig.DbName)
	default:
		errorString := "Cannot connect. Unsupported DB type: (" + dbConfig.DbType + ")"
		log.Println(errorString)
		return nil, errors.New(errorString)
	}
	if err != nil {
		log.Println("Error!")
		return nil, err
	}

	if err = DB.DB().Ping(); err != nil {
		log.Println("Unable to verify connection to database")
		return nil, err
	}
	return &DB, nil
}

// InternalDBInit initializes the internal database connection that the service broker will use.
// In addition to calling DBInit(), it also makes sure that the tables are setup for Instance and DBConfig structs.
func InternalDBInit(dbConfig *common.DBConfig) (*gorm.DB, error) {
	db, err := DBInit(dbConfig)
	if err == nil {
		db.DB().SetMaxOpenConns(10)
		log.Println("Migrating")
		db.LogMode(true)
		// Automigrate!
		db.AutoMigrate(&rds.RDSInstance{}, &base.Instance{}) // Add all your models here to help setup the database tables
		log.Println("Migrated")
	}
	return db, err
}
