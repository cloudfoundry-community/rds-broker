package redis

import (
	"github.com/18F/aws-broker/base"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/jinzhu/gorm"

	"github.com/18F/aws-broker/catalog"
	"github.com/18F/aws-broker/config"

	"fmt"
	"log"
)

type redisAdapter interface {
	createRedis(i *RedisInstance, password string) (base.InstanceState, error)
	bindRedisToApp(i *RedisInstance, password string) (map[string]string, error)
	deleteRedis(i *RedisInstance) (base.InstanceState, error)
}

type sharedRedisAdapter struct {
	SharedRedisConn *gorm.DB
}

func (d *sharedRedisAdapter) createDB(i *RedisInstance, password string) (base.InstanceState, error) {
	return base.InstanceReady, nil
}

func (d *sharedRedisAdapter) bindDBToApp(i *RedisInstance, password string) (map[string]string, error) {
	return i.getCredentials(password)
}

func (d *sharedRedisAdapter) deleteRedis(i *RedisInstance) (base.InstanceState, error) {
	return base.InstanceGone, nil
}

type dedicatedRedisAdapter struct {
	Plan     catalog.RDSPlan
	settings config.Settings
}

// This is the prefix for all pgroups created by the broker.
const PgroupPrefix = "cg-redis-broker-"

func (d *dedicatedRedisAdapter) createRedis(i *RedisInstance, password string) (base.InstanceState, error) {
	svc := elasticache.New(session.New(), aws.NewConfig().WithRegion(d.settings.Region))
	var redisTags []*elasticache.Tag

	for k, v := range i.Tags {
		var tag elasticache.Tag
		tag = elasticache.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		redisTags = append(redisTags, &tag)
	}

	// Standard parameters
	params := &elasticache.CreateCacheClusterInput{
		AutoMinorVersionUpgrade:   aws.Bool(true),
		CacheClusterId:            aws.String("aws-broker-redis-test"),
		CacheNodeType:             aws.String("cache.t3.micro"),
		CacheSubnetGroupName:      aws.String("default"),
		Engine:                    aws.String("redis"),
		EngineVersion:             aws.String("5.0.6"),
		NumCacheNodes:             aws.Int64(1),
		Port:                      aws.Int64(6379),
		PreferredAvailabilityZone: aws.String("us-gov-west-1"),
		SnapshotRetentionLimit:    aws.Int64(7),
	}

	resp, err := svc.CreateCacheCluster(params)
	// Pretty-print the response data.
	log.Println(awsutil.StringValue(resp))
	// Decide if AWS service call was successful
	if yes := d.didAwsCallSucceed(err); yes {
		return base.InstanceInProgress, nil
	}
	return base.InstanceNotCreated, nil
}

func (d *dedicatedRedisAdapter) bindRedisToApp(i *RedisInstance, password string) (map[string]string, error) {
	// First, we need to check if the instance is up and available before binding.
	// Only search for details if the instance was not indicated as ready.
	if i.State != base.InstanceReady {
		svc := elasticache.New(session.New(), aws.NewConfig().WithRegion(d.settings.Region))
		params := &elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String("aws-broker-redis-test"), // Required
		}

		resp, err := svc.DescribeCacheClusters(params)
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
	}
	// If we get here that means the instance is up and we have the information for it.
	return i.getCredentials(password)
}

func (d *dedicatedRedisAdapter) deleteDB(i *RedisInstance) (base.InstanceState, error) {
	svc := elasticache.New(session.New(), aws.NewConfig().WithRegion(d.settings.Region))
	params := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String("aws-broker-redis-test"), // Required
	}
	resp, err := svc.DeleteCacheCluster(params)
	// Pretty-print the response data.
	fmt.Println(awsutil.StringValue(resp))

	// Decide if AWS service call was successful
	if yes := d.didAwsCallSucceed(err); yes {
		return base.InstanceGone, nil
	}
	return base.InstanceNotGone, nil
}

func (d *dedicatedRedisAdapter) didAwsCallSucceed(err error) bool {
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
