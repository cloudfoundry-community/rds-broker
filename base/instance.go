package base

import (
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
	Id   int64
	Uuid string `sql:"size(255)"`

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

	State InstanceState

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
