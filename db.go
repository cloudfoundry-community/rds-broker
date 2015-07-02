package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"errors"
	"fmt"
	"log"
)

// DBConfig holds configuration information to connect to a database.
// Parameters for the config.
// * dbname - The name of the database to connect to
// * user - The user to sign in as
// * password - The user's password
// * host - The host to connect to. Values that start with / are for unix domain sockets.
//   (default is localhost)
// * port - The port to bind to. (default is 5432)
// * sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
//   Valid SSL modes:
//    * disable - No SSL
//    * require - Always SSL (skip verification)
//    * verify-full - Always SSL (require verification)
type DBConfig struct {
	DbType   string
	Url      string
	Username string
	Password string
	DbName   string
	Sslmode  string
	Port     string
}

// DBinit is a generic helper function that will try to connect to a database with the config in the input.
// Supported DB types:
// * postgres
// * sqlite3
func DBInit(dbConfig *DBConfig) (*gorm.DB, error) {
	var DB gorm.DB
	var err error
	switch dbConfig.DbType {
	case "postgres":
		conn := "dbname=%s user=%s password=%s host=%s sslmode=%s port=%s"
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
func InternalDBInit(dbConfig *DBConfig) (*gorm.DB, error) {
	db, err := DBInit(dbConfig)
	if err == nil {
		db.DB().SetMaxOpenConns(10)
		log.Println("Migrating")
		// Automigrate!
		db.AutoMigrate(Instance{}) // Add all your models here to help setup the database tables.
		log.Println("Migrated")
	}
	return db, err
}
