package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
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
		PubliclyAccessible:      aws.Boolean(true),
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

	// FIXME.
	if *params.DBInstanceClass == "db.t2.micro" {
		// FIXME PART 1.
		// A micro instance can support Multi-AZ however it has to be in a VPC. Since right now we aren't
		// configuring VPCs, we will just turn off Multi-AZ for a micro instance.
		params.MultiAZ = aws.Boolean(false)

		// FIXME PART 2.
		// A micro instance needs to be apart of a VPC.
		var subnetID *string
		var vpcID *string
		var subnet *ec2.Subnet
		ec2Svc := ec2.New(&aws.Config{Region: "us-east-1"})
		// Get the DB Subnet Group Name.

		// Get the EC2 Subnets and VPCIds
		describeSubnetparams := &ec2.DescribeSubnetsInput{
		// TODO: Add parameters.
		}
		describeSubnetsResp, err := ec2Svc.DescribeSubnets(describeSubnetparams)
		fmt.Println(awsutil.StringValue(describeSubnetsResp))
		if !d.DidAwsCallSucceed(err) || len(describeSubnetsResp.Subnets) < 1 {
			// If no subnet exists, create VPC and then create subnet.
			// TODO
			/*
				cidrBlock = "10.0.0.0/16"
				ec2VPCParams := &ec2.CreateVPCInput{
					CIDRBlock:       aws.String(cidrBlock), // Required.
					InstanceTenancy: aws.String("default"),
				}
				ec2VPCResp, err := ec2Svc.CreateVPC(ec2VPCParams)
				fmt.Println(awsutil.StringValue(ec2VPCResp))
				if !d.DidAwsCallSucceed(err) {
					return InstanceNotCreated, nil
				}
				vpcID = ec2VPCResp.VPC.VPCID
				// Create a subnet with the VPC just created.
				ec2SubnetParams := &ec2.CreateSubnetInput{
					CIDRBlock: aws.String(cidrBlock), // Required
					VPCID:     vpcID,                 // Required
				}
				ec2SubnetResp, err := ec2Svc.CreateSubnet(ec2SubnetParams)
				fmt.Println(awsutil.StringValue(ec2SubnetResp))
				if !d.DidAwsCallSucceed(err) {
					return InstanceNotCreated, nil
				}
			*/
			return InstanceNotCreated, nil
		} else {
			// Just get the vpcid and cidrblock of the first one.
			subnet = describeSubnetsResp.Subnets[0]
			vpcID = subnet.VPCID
		}
		// Look for a db subnet created with our VPC.
		describeDbSubnetGroupsParams := &rds.DescribeDBSubnetGroupsInput{
		// TODO When AWS SDK supports filtering (when it does, we will filter by vpc or subnetid).
		}
		_ = vpcID // Temp FIXME until filtering.

		// Find exisiting db subnet group names.
		describeDbSubnetGroupsResp, err := svc.DescribeDBSubnetGroups(describeDbSubnetGroupsParams)
		fmt.Println(awsutil.StringValue(describeDbSubnetGroupsResp))
		if !d.DidAwsCallSucceed(err) || len(describeDbSubnetGroupsResp.DBSubnetGroups) < 1 {
			// Create db subnet group with VPC if a DB Subnet Group does not exist.
			createDBSubnetGroupParams := &rds.CreateDBSubnetGroupInput{
				DBSubnetGroupDescription: aws.String(i.PlanId + "-subnet-" + randStr(10)),
				DBSubnetGroupName:        aws.String(randStr(10)),
				SubnetIDs:                []*string{subnet.SubnetID},
			}
			createDBSubnetGroupResp, err := svc.CreateDBSubnetGroup(createDBSubnetGroupParams)
			if !d.DidAwsCallSucceed(err) {
				return InstanceNotCreated, nil
			}
			// Assign db subnet group name
			subnetID = createDBSubnetGroupResp.DBSubnetGroup.DBSubnetGroupName
		} else {
			// Assign db subnet group name
			subnetID = describeDbSubnetGroupsResp.DBSubnetGroups[0].DBSubnetGroupName
		}

		// Assign Subnet group name into create instance parameters.
		params.DBSubnetGroupName = subnetID
	}
	// END FIXME

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

func (d *DedicatedDBAdapter) BindDBToApp(i *Instance, password string) (map[string]string, error) {
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
