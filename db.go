package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"log"
	"errors"
)

// Connection string parameters for Postgres - http://godoc.org/github.com/lib/pq, if you are using another
// database refer to the relevant driver's documentation.

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

var DB gorm.DB

func DBInit(rds *RDS) error {
	var err error
	switch rds.DbType {
	case "postgres":
		DB, err = gorm.Open("postgres", "user=gorm dbname=gorm sslmode=disable")
	case "sqlite3":
		DB, err = gorm.Open("sqlite3", rds.DbName)
	default:
		errorString := "Cannot connect. Unsupported DB type: (" + rds.DbType + ")"
		log.Println(errorString)
		return errors.New(errorString)
	}
	if err != nil {
		log.Println("Error!")
		return err
	}

	if err = DB.DB().Ping(); err != nil {
		log.Println("Unable to verify connection to database")
		return err
	}
	DB.DB().SetMaxOpenConns(10)
	log.Println("Migrating")
	// Automigrate!
	DB.AutoMigrate(Instance{})
	log.Println("Migrated")
	return nil
}
