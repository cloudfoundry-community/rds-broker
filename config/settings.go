package config

import (
	"errors"
	"fmt"
	"github.com/18F/aws-broker/common"
	"log"
	"os"
	"strconv"
)

// Settings stores settings used to run the application
type Settings struct {
	EncryptionKey string
	DbConfig      *common.DBConfig
	Environment   string
}

// LoadFromEnv loads settings from environment variables
func (s *Settings) LoadFromEnv() error {
	log.Println("Loading settings")

	// Load DB Settings
	dbConfig := common.DBConfig{}
	dbConfig.DbType = os.Getenv("DB_TYPE")
	dbConfig.URL = os.Getenv("DB_URL")
	dbConfig.Username = os.Getenv("DB_USER")
	dbConfig.Password = os.Getenv("DB_PASS")
	dbConfig.DbName = os.Getenv("DB_NAME")
	if dbConfig.Sslmode = os.Getenv("DB_SSLMODE"); dbConfig.Sslmode == "" {
		dbConfig.Sslmode = "require"
	}

	// Ensure AWS credentials exist in environment
	for _, key := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_DEFAULT_REGION"} {
		if os.Getenv(key) == "" {
			return fmt.Errorf("Must set environment variable %s", key)
		}
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

	// Set env to production
	s.Environment = "production"

	return nil
}
