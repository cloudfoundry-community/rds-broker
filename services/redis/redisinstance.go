package redis

import (
	"fmt"
	"strconv"

	"github.com/18F/aws-broker/base"

	"github.com/18F/aws-broker/catalog"
	"github.com/18F/aws-broker/config"
)

// RedisInstance represents the information of a Redis Service instance.
type RedisInstance struct {
	base.Instance

	Username string `sql:"size(255)"`
	Password string `sql:"size(255)"`

	Tags            map[string]string `sql:"-"`
	DbSubnetGroup   string            `sql:"-"`
	SecGroup        string            `sql:"-"`
	EnableFunctions bool              `sql:"-"`
}

func (i *RedisInstance) setPassword(password, key string) error {
	i.Password = password
	return nil
}

func (i *RedisInstance) getPassword(key string) (string, error) {
	return i.Password, nil
}

func (i *RedisInstance) getCredentials(password string) (map[string]string, error) {
	var credentials map[string]string

	uri := fmt.Sprintf("redis://%s:%s@%s:%d",
		i.Username,
		password,
		i.Host,
		i.Port)

	credentials = map[string]string{
		"uri":      uri,
		"username": i.Username,
		"password": password,
		"host":     i.Host,
		"port":     strconv.FormatInt(i.Port, 10),
	}
	return credentials, nil
}

func (i *RedisInstance) init(uuid string,
	orgGUID string,
	spaceGUID string,
	serviceID string,
	plan catalog.RedisPlan,
	options RedisOptions,
	s *config.Settings) error {

	i.Uuid = uuid
	i.ServiceID = serviceID
	i.PlanID = plan.ID
	i.OrganizationGUID = orgGUID
	i.SpaceGUID = spaceGUID

	// Load AWS values
	i.DbSubnetGroup = plan.SubnetGroup
	i.SecGroup = plan.SecurityGroup

	// Load tags
	i.Tags = plan.Tags

	// Tag instance with broker details
	i.Tags["Instance GUID"] = uuid
	i.Tags["Space GUID"] = spaceGUID
	i.Tags["Organization GUID"] = orgGUID
	i.Tags["Plan GUID"] = plan.ID
	i.Tags["Service GUID"] = serviceID

	return nil
}
