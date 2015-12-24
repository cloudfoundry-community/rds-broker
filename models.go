package main

import (
	// "github.com/jinzhu/gorm"
	// _ "github.com/lib/pq"

	"crypto/aes"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type Instance struct {
	Id       int64
	Uuid     string `sql:"size(255)"`
	Database string `sql:"size(255)"`
	Username string `sql:"size(255)"`
	Password string `sql:"size(255)"`
	Salt     string `sql:"size(255)"`

	ClearPassword string `sql:"-"`

	ServiceId string `sql:"size(255)"`
	PlanId    string `sql:"size(255)"`
	OrgGuid   string `sql:"size(255)"`
	SpaceGuid string `sql:"size(255)"`

	Tags          map[string]string `sql:"-"`
	DbSubnetGroup string            `sql:"-"`
	SecGroup      string            `sql:"-"`

	Adapter string `sql:"size(255)"`

	Host string `sql:"size(255)"`
	Port int64

	DbType string `sql:"size(255)"`

	State DBInstanceState

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (i *Instance) SetPassword(password, key string) error {
	if i.Salt == "" {
		return errors.New("Salt has to be set before writing the password")
	}

	iv, _ := base64.StdEncoding.DecodeString(i.Salt)

	encrypted, err := Encrypt(password, key, iv)
	if err != nil {
		return err
	}

	i.Password = encrypted
	i.ClearPassword = password

	return nil
}

func (i *Instance) GetPassword(key string) (string, error) {
	if i.Salt == "" || i.Password == "" {
		return "", errors.New("Salt and password has to be set before writing the password")
	}

	iv, _ := base64.StdEncoding.DecodeString(i.Salt)

	decrypted, err := Decrypt(i.Password, key, iv)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

func (i *Instance) GetCredentials(password string) (map[string]string, error) {
	var credentials map[string]string
	switch i.DbType {
	case "postgres":
		uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			i.Username,
			password,
			i.Host,
			i.Port,
			i.Database)

		credentials = map[string]string{
			"uri":      uri,
			"username": i.Username,
			"password": password,
			"host":     i.Host,
			"db_name":  i.Database,
		}
	default:
		return nil, errors.New("Cannot generate credentials for unsupported db type: " + i.DbType)
	}
	return credentials, nil
}

func (i *Instance) Init(uuid string,
	orgGuid string,
	spaceGuid string,
	serviceId string,
	plan *AWSPlan,
	s *Settings) error {

	i.Uuid = uuid
	i.ServiceId = serviceId
	i.PlanId = plan.ID
	i.OrgGuid = orgGuid
	i.SpaceGuid = spaceGuid

	i.Adapter = plan.Adapter

	// Build random values
	i.Database = "db" + randStr(15)
	i.Username = "u" + randStr(15)
	i.Salt = GenerateSalt(aes.BlockSize)
	password := randStr(25)
	if err := i.SetPassword(password, s.EncryptionKey); err != nil {
		return err
	}

	// Load tags
	i.Tags = s.InstanceTags

	// Load AWS values
	i.DbType = plan.DbType
	i.DbSubnetGroup = s.SubnetGroup
	i.SecGroup = s.SecGroup

	return nil
}
