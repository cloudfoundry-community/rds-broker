package main

import (
	// "github.com/jinzhu/gorm"
	// _ "github.com/lib/pq"

	"encoding/base64"
	"errors"
	"time"
)

type Instance struct {
	Id       int64
	Uuid     string `sql:"size(255)"`
	Database string `sql:"size(255)"`
	Username string `sql:"size(255)"`
	Password string `sql:"size(255)"`
	Salt     string `sql:"size(255)"`

	PlanId    string `sql:"size(255)"`
	OrgGuid   string `sql:"size(255)"`
	SpaceGuid string `sql:"size(255)"`

	Tags          map[string]string `sql:"-"`
	DBSubnetGroup string            `sql:"-"`

	Adapter      string `sql:"size(255)"`

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
