package main

import (
	"github.com/jinzhu/gorm"

	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
)

// Settings stores settings used to run the application
type Settings struct {
	EncryptionKey string
	DbConfig      *DBConfig
	InstanceTags  map[string]string
	Environment   string
	SecGroup      string
	SubnetGroup   string
}

// InitializeAdapter is the main function to create database instances
func (s Settings) InitializeAdapter(plan *AWSPlan,
	sharedDbConn *gorm.DB) (DBAdapter, error) {

	var dbAdapter DBAdapter
	// For test environments, use a mock adapter.
	if s.Environment == "test" {
		dbAdapter = &MockDBAdapter{}
		return dbAdapter, nil
	}

	switch plan.Adapter {
	case "shared":
		dbAdapter = &SharedDBAdapter{
			SharedDbConn: sharedDbConn,
		}
	case "dedicated":
		dbAdapter = &DedicatedDBAdapter{
			InstanceType: plan.InstanceType,
		}
	default:
		return nil, errors.New("Adapter not found")
	}

	return dbAdapter, nil
}

// LoadFromEnv loads settings from environment variables
func (s *Settings) LoadFromEnv() error {
	log.Println("Loading settings")

	// Load DB Settings
	dbConfig := DBConfig{}
	dbConfig.DbType = os.Getenv("DB_TYPE")
	dbConfig.Url = os.Getenv("DB_URL")
	dbConfig.Username = os.Getenv("DB_USER")
	dbConfig.Password = os.Getenv("DB_PASS")
	dbConfig.DbName = os.Getenv("DB_NAME")
	if dbConfig.Sslmode = os.Getenv("DB_SSLMODE"); dbConfig.Sslmode == "" {
		dbConfig.Sslmode = "require"
	}

	if os.Getenv("DB_PORT") != "" {
		var err error
		dbConfig.Port, err = strconv.ParseInt(os.Getenv("DB_PORT"), 10, 64)
		// Just return nothing if we can't interpret the number.
		if err != nil {
			return errors.New("Couldn't load port number")
		}
	} else {
		dbConfig.Port = 5432
	}

	s.DbConfig = &dbConfig

	// Load Encryption Key
	s.EncryptionKey = os.Getenv("ENC_KEY")
	if s.EncryptionKey == "" {
		return errors.New("An encryption key is required")
	}

	// Load tags
	tags := os.Getenv("INSTANCE_TAGS")
	if tags != "" {
		json.Unmarshal([]byte(tags), &s.InstanceTags)
	}

	// Load AWS settings
	s.SecGroup = os.Getenv("AWS_SEC_GROUP")
	s.SubnetGroup = os.Getenv("AWS_DB_SUBNET_GROUP")

	// Set env to production
	s.Environment = "production"

	return nil

}
