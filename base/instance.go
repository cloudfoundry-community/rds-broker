package base

import (
	"github.com/cloudfoundry-community/aws-broker/helpers/request"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"time"
)

type InstanceState uint8

const (
	InstanceNotCreated InstanceState = iota // 0
	InstanceInProgress                      // 1
	InstanceReady                           // 2
	InstanceGone                            // 3
	InstanceNotGone                         // 4
)

type Instance struct {
	Id   string `gorm:"primary_key" sql:"type:varchar(255) PRIMARY KEY"`
	Uuid string `sql:"size(255)"`

	request.Request

	Host string `sql:"size(255)"`
	Port int64

	State InstanceState

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func FindBaseInstance(brokerDb *gorm.DB, id string) (Instance, response.Response) {
	instance := Instance{}
	log.Println("Looking for instance with id " + id)
	err := brokerDb.Where("id = ?", id).First(&instance).Error
	if err == nil {
		return instance, nil
	} else if err == gorm.RecordNotFound {
		return instance, response.NewErrorResponse(http.StatusNotFound, err.Error())
	} else {
		return instance, response.NewErrorResponse(http.StatusInternalServerError, err.Error())
	}
}
