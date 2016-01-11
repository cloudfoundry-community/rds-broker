package db

import (
	"github.com/cloudfoundry-community/aws-broker/base"
	"github.com/cloudfoundry-community/aws-broker/common"
	"github.com/cloudfoundry-community/aws-broker/services/rds"
	"github.com/jinzhu/gorm"
	"log"
)

// InternalDBInit initializes the internal database connection that the service broker will use.
// In addition to calling DBInit(), it also makes sure that the tables are setup for Instance and DBConfig structs.
func InternalDBInit(dbConfig *common.DBConfig) (*gorm.DB, error) {
	db, err := common.DBInit(dbConfig)
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
