package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/jinzhu/gorm"

	"errors"
	"fmt"
	"log"
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
	CreateDB(i *Instance, password string) (DBInstanceState, error)
	BindDBToApp(i *Instance, password string) (map[string]string, error)
	DeleteDB(i *Instance) (DBInstanceState, error)
}

// MockDBAdapter is a struct meant for testing.
// It should only be used in *_test.go files.
// It is only here because *_test.go files are only compiled during "go test"
// and it's referenced in non *_test.go code eg. InitializeAdapter in main.go.
type MockDBAdapter struct {
}

func (d *MockDBAdapter) CreateDB(i *Instance, password string) (DBInstanceState, error) {
	// TODO
	return InstanceReady, nil
}

func (d *MockDBAdapter) BindDBToApp(i *Instance, password string) (map[string]string, error) {
	// TODO
	return i.GetCredentials(password)
}

func (d *MockDBAdapter) DeleteDB(i *Instance) (DBInstanceState, error) {
	// TODO
	return InstanceGone, nil
}

// END MockDBAdpater

type SharedDBAdapter struct {
	SharedDbConn *gorm.DB
}

func (d *SharedDBAdapter) CreateDB(i *Instance, password string) (DBInstanceState, error) {
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

func (d *SharedDBAdapter) BindDBToApp(i *Instance, password string) (map[string]string, error) {
	return i.GetCredentials(password)
}

func (d *SharedDBAdapter) DeleteDB(i *Instance) (DBInstanceState, error) {
	if db := d.SharedDbConn.Exec(fmt.Sprintf("DROP DATABASE %s;", i.Database)); db.Error != nil {
		return InstanceNotGone, db.Error
	}
	if db := d.SharedDbConn.Exec(fmt.Sprintf("DROP USER %s;", i.Username)); db.Error != nil {
		return InstanceNotGone, db.Error
	}
	return InstanceGone, nil
}

type DedicatedDBAdapter struct {
	InstanceType string
}

func (d *DedicatedDBAdapter) CreateDB(i *Instance, password string) (DBInstanceState, error) {
	svc := rds.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	var rdsTags []*rds.Tag

	for k, v := range i.Tags {
		var tag rds.Tag
		tag = rds.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		rdsTags = append(rdsTags, &tag)
	}

	// Standard parameters
	params := &rds.CreateDBInstanceInput{
		// Everyone gets 10gb for now
		AllocatedStorage: aws.Int64(10),
		// Instance class is defined by the plan
		DBInstanceClass:         &d.InstanceType,
		DBInstanceIdentifier:    &i.Database,
		DBName:                  &i.Database,
		Engine:                  aws.String("postgres"),
		MasterUserPassword:      &password,
		MasterUsername:          &i.Username,
		AutoMinorVersionUpgrade: aws.Bool(true),
		MultiAZ:                 aws.Bool(true),
		StorageEncrypted:        aws.Bool(true),
		Tags:                    rdsTags,
		PubliclyAccessible:      aws.Bool(false),
		DBSubnetGroupName:       &i.DbSubnetGroup,
		VpcSecurityGroupIds: []*string{
			aws.String("String"),
			&i.SecGroup,
		},
	}

	if *params.DBInstanceClass == "db.t2.micro" {
		params.StorageEncrypted = aws.Bool(false)
	}

	resp, err := svc.CreateDBInstance(params)
	// Pretty-print the response data.
	log.Println(awsutil.StringValue(resp))
	// Decide if AWS service call was successful
	if yes := d.DidAwsCallSucceed(err); yes {
		return InstanceInProgress, nil
	} else {
		return InstanceNotCreated, nil
	}
}

func (d *DedicatedDBAdapter) BindDBToApp(i *Instance, password string) (map[string]string, error) {
	// First, we need to check if the instance is up and available before binding.
	// Only search for details if the instance was not indicated as ready.
	if i.State != InstanceReady {
		svc := rds.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
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
					return nil, errors.New("Instance not available yet. Please wait and try again..")
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

func (d *DedicatedDBAdapter) DeleteDB(i *Instance) (DBInstanceState, error) {
	svc := rds.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(i.Database), // Required
		// FinalDBSnapshotIdentifier: aws.String("String"),
		SkipFinalSnapshot: aws.Bool(true),
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

func (d *DedicatedDBAdapter) DidAwsCallSucceed(err error) bool {
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
