package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
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
	InstanceGone                              // 3
	InstanceNotGone                           // 4
)

type DBAdapter interface {
	CreateDB(plan *Plan, db *gorm.DB) (*DB, error)
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
	sharedDbConn *gorm.DB) (*DB, error) {

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
		return nil, errors.New("Adapter not found")
	}

	return &db, nil
}

type DB interface {
	CreateDB(i *Instance, password string) (DBInstanceState, error)
	BindDBToApp(i *Instance, password string) (map[string]string, error)
	DeleteDB(i *Instance) (DBInstanceState, error)
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

func (d *SharedDB) BindDBToApp(i *Instance, password string) (map[string]string, error) {
	return i.GetCredentials(password)
}

func (d * SharedDB) DeleteDB(i *Instance) (DBInstanceState, error) {
	if db := d.SharedDbConn.Exec(fmt.Sprintf("DROP DATABASE %s;", i.Database)); db.Error != nil {
		return InstanceNotGone, db.Error
	}
	if db := d.SharedDbConn.Exec(fmt.Sprintf("DROP USER %s;", i.Username)); db.Error != nil {
		return InstanceNotGone, db.Error
	}
	return InstanceGone, nil
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

	// Standard parameters
	params := &rds.CreateDBInstanceInput{
		// Everyone gets 10gb for now
		AllocatedStorage: aws.Long(10),
		// Instance class is defined by the plan
		DBInstanceClass:         &d.InstanceType,
		DBInstanceIdentifier:    &i.Database,
		Engine:                  aws.String("postgres"),
		MasterUserPassword:      &password,
		MasterUsername:          &i.Username,
		AutoMinorVersionUpgrade: aws.Boolean(true),
		MultiAZ:                 aws.Boolean(true),
		StorageEncrypted:        aws.Boolean(true),
		Tags:                    rdsTags,
	}

	// Now, adjust parameters based on the particular instance.
	// Per http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.Encryption.html,
	// Encryption is only supported on the following instances.
	switch *params.DBInstanceClass {
	// Start list of supported instance types for encryption.
	case "db.m3.medium":
		fallthrough
	case "db.m3.large":
		fallthrough
	case "db.m3.xlarge":
		fallthrough
	case "db.m3.2xlarge":
		fallthrough
	case "db.r3.large":
		fallthrough
	case "db.r3.xlarge":
		fallthrough
	case "db.r3.2xlarge":
		fallthrough
	case "db.r3.4xlarge":
		fallthrough
	case "db.r3.8xlarge":
		fallthrough
	case "db.cr1.8xlarge":
		// End of supported instance types.
		_ = 0
	default:
		fmt.Println("Encryption not supported by AWS for instance size: " + d.InstanceType)
		params.StorageEncrypted = aws.Boolean(false)
	}

	resp, err := svc.CreateDBInstance(params)
	// Pretty-print the response data.
	fmt.Println(awsutil.StringValue(resp))
	// Decide if AWS service call was successful
	if yes := d.DidAwsCallSucceed(err); yes {
		return InstanceInProgress, nil
	} else {
		return InstanceNotCreated, nil
	}
}

func (d *DedicatedDB) BindDBToApp(i *Instance, password string) (map[string]string, error) {
	// First, we need to check if the instance is up and available before binding.
	// Only search for details if the instance was not indicated as ready.
	if i.State != InstanceReady {
		svc := rds.New(&aws.Config{Region: "us-east-1"})
		params := &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(i.Database),
			// MaxRecords: aws.Long(1),
		}

		resp, err := svc.DescribeDBInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// Generic AWS error with Code, Message, and original error (if any)
				fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					// A service error occurred
					fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				// This case should never be hit, the SDK should always return an
				// error which satisfies the awserr.Error interface.
				fmt.Println(err.Error())
			}
			return nil, err
		}

		// Pretty-print the response data.
		fmt.Println(awsutil.StringValue(resp))

		// Get the details (host and port) for the instance.
		numOfInstances := len(resp.DBInstances)
		if numOfInstances > 0 {
			for _, value := range resp.DBInstances {
				// First check that the instance is up.
				if value.DBInstanceStatus != nil && *(value.DBInstanceStatus) == "available" {
					if value.Endpoint != nil && value.Endpoint.Address != nil && value.Endpoint.Port != nil {
						fmt.Printf("host: %s port: %d \n", *(value.Endpoint.Address), *(value.Endpoint.Port))
						i.Port = *(value.Endpoint.Port)
						i.Host = *(value.Endpoint.Address)
						i.State = InstanceReady
						// Should only be one regardless. Just return now.
						break
					} else {
						// Something went horribly wrong. Should never get here.
						return nil, errors.New("Inavlid memory for endpoint and/or endpoint members.")
					}
				} else {
					// Instance not up yet.
					return nil, errors.New("Instance not available.")
				}
			}
		} else {
			// Couldn't find any instances.
			return nil, errors.New("Couldn't find any instances.")
		}
	}
	// If we get here that means the instance is up and we have the information for it.
	return i.GetCredentials(password)
}

func (d *DedicatedDB) DeleteDB(i *Instance) (DBInstanceState, error) {
	svc := rds.New(&aws.Config{Region: "us-east-1"})
	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(i.Database), // Required
		// FinalDBSnapshotIdentifier: aws.String("String"),
		SkipFinalSnapshot: aws.Boolean(true),
	}
	resp, err := svc.DeleteDBInstance(params)
	// Pretty-print the response data.
	fmt.Println(awsutil.StringValue(resp))
	// Decide if AWS service call was successful
	if yes := d.DidAwsCallSucceed(err); yes {
		return InstanceGone, nil
	} else {
		return InstanceNotGone, nil
	}
}

func (d *DedicatedDB) DidAwsCallSucceed(err error) bool {
	// TODO Eventually return a formatted error object.
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Generic AWS Error with Code, Message, and original error (if any)
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			// This case should never be hit, The SDK should alwsy return an
			// error which satisfies the awserr.Error interface.
			fmt.Println(err.Error())
		}
		return false
	}
	return true
}
