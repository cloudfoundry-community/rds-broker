package main

import (
	// "github.com/jinzhu/gorm"
	// _ "github.com/lib/pq"

	"time"
)

type Instance struct {
	Id       int64
	Uuid     string `sql:"size(255)"`
	Database string `sql:"size(255)"`
	Username string `sql:"size(255)"`
	Password string `sql:"size(255)"`

	PlanId    string `sql:"size(255)"`
	OrgGuid   string `sql:"size(255)"`
	SpaceGuid string `sql:"size(255)"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
