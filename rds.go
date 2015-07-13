package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/jinzhu/gorm"

	"errors"
	"fmt"
)

type DBInstanceState uint8

const (
	InstanceNotCreated DBInstanceState = iota // 0
	InstanceInProgress                        // 1
	InstanceReady                             // 2
)

type DBAdapter interface {
	CreateDB(plan *Plan, i *Instance, db *gorm.DB, password string) (DBInstanceState, error)
}

type RDSAdapter struct {
}

// Main function to create database instances
// Selects an adapter and depending on the plan
// creates the instance
// Returns status and error
// Status codes:
// 0 = not created
// 1 = in progress
// 2 = ready
func (a RDSAdapter) CreateDB(plan *Plan,
	i *Instance,
	sharedDbConn *gorm.DB,
	password string) (DBInstanceState, error) {

	var db DB
	switch plan.Adapter {
	case "shared":
		db = &SharedDB{
			SharedDbConn: sharedDbConn,
		}
	case "dedicated":
		db = &DedicatedDB{
			InstanceType: plan.InstanceType,
		}
	default:
		return InstanceNotCreated, errors.New("Adapter not found")
	}

	status, err := db.CreateDB(i, password)
	return status, err
}

type DB interface {
	CreateDB(i *Instance, password string) (DBInstanceState, error)
}

type SharedDB struct {
	SharedDbConn *gorm.DB
}

func (d *SharedDB) CreateDB(i *Instance, password string) (DBInstanceState, error) {
	if db := d.SharedDbConn.Exec(fmt.Sprintf("CREATE DATABASE %s;", i.Database)); db.Error != nil {
		return InstanceNotCreated, db.Error
	}
	if db := d.SharedDbConn.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", i.Username, password)); db.Error != nil {
		// TODO. Revert CREATE DATABASE.
		return InstanceNotCreated, db.Error
	}
	if db := d.SharedDbConn.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", i.Database, i.Username)); db.Error != nil {
		// TODO. Revert CREATE DATABASE and CREATE USER.
		return InstanceNotCreated, db.Error
	}
	return InstanceReady, nil
}

type DedicatedDB struct {
	InstanceType string
}

func (d *DedicatedDB) CreateDB(i *Instance, password string) (DBInstanceState, error) {
	svc := rds.New(&aws.Config{Region: "us-east-1"})

	var rdsTags []*rds.Tag

	for k, v := range i.Tags {
		rdsTags = append(rdsTags, &rds.Tag{
			Key:   &k,
			Value: &v,
		})
	}

	params := &rds.CreateDBInstanceInput{
		// Everyone gets 10gb for now
		AllocatedStorage: aws.Long(10),
		// Instance class is defined by the plan
		DBInstanceClass:         &d.InstanceType,
		DBInstanceIdentifier:    &i.Database,
		Engine:                  aws.String("postgres"),
		MasterUserPassword:      &i.Password,
		MasterUsername:          &i.Username,
		AutoMinorVersionUpgrade: aws.Boolean(true),
		DBSecurityGroups: []*string{
			aws.String("String"), // Required
			// More values...
		},
		DBSubnetGroupName: aws.String("String"),
		MultiAZ:           aws.Boolean(true),
		StorageEncrypted:  aws.Boolean(true),
		Tags:              rdsTags,
		VPCSecurityGroupIDs: []*string{
			aws.String("String"), // Required
			// More values...
		},
	}
	resp, err := svc.CreateDBInstance(params)

	_ = resp
	_ = err

	// if err != nil {
	// 	if awsErr, ok := err.(awserr.Error); ok {
	// 		// Generic AWS Error with Code, Message, and original error (if any)
	// 		fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
	// 		if reqErr, ok := err.(awserr.RequestFailure); ok {
	// 			// A service error occurred
	// 			fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
	// 		}
	// 	} else {
	// 		// This case should never be hit, The SDK should alwsy return an
	// 		// error which satisfies the awserr.Error interface.
	// 		fmt.Println(err.Error())
	// 	}
	// }

	// // Pretty-print the response data.
	// fmt.Println(awsutil.StringValue(resp))

	return InstanceNotCreated, nil
}
