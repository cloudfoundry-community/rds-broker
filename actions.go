package main

import (
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	"crypto/aes"
	"fmt"
)

type Response struct {
	Description string `json: description`
}

// CreateInstance
// URL: /v2/service_instances/:id
// Params:
// {
//   "service_id":        "service-guid-here",
//   "plan_id":           "plan-guid-here",
//   "organization_guid": "org-guid-here",
//   "space_guid":        "space-guid-here"
// }
func CreateInstance(p martini.Params, r render.Render, db *gorm.DB, s *Settings) {
	instance := Instance{}

	db.Where("uuid = ?", p["id"]).First(&instance)

	if instance.Id > 0 {
		r.JSON(409, Response{"The instance already exists"})
		return
	}

	instance.Uuid = p["id"]
	instance.PlanId = p["plan_id"]
	instance.OrgGuid = p["organization_guid"]
	instance.SpaceGuid = p["space_guid"]

	instance.Database = "db" + randStr(15)
	instance.Username = "u" + randStr(15)
	instance.Salt = GenerateSalt(aes.BlockSize)
	password := randStr(25)
	err := instance.SetPassword(password, s.EncryptionKey)
	if err != nil {
		desc := "There was an error setting the password" + err.Error()
		r.JSON(500, Response{desc})
		return
	}

	// Create the database
	db.Exec(fmt.Sprintf("CREATE DATABASE %s;", instance.Database))
	db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", instance.Username, password))
	db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", instance.Database, instance.Username))

	db.Save(&instance)

	r.JSON(201, Response{"The instance was created"})
}

// BindInstance
// URL: /v2/service_instances/:instance_id/service_bindings/:binding_id
// Params:
// {
//   "plan_id":        "plan-guid-here",
//   "service_id":     "service-guid-here",
//   "app_guid":       "app-guid-here"
// }
func BindInstance(p martini.Params, r render.Render, db *gorm.DB, s *Settings) {
	instance := Instance{}

	db.Where("uuid = ?", p["instance_id"]).First(&instance)
	if instance.Id == 0 {
		r.JSON(404, Response{"Instance not found"})
		return
	}

	password, err := instance.GetPassword(s.EncryptionKey)
	if err != nil {
		r.JSON(500, "")
	}

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		instance.Username,
		password,
		s.Rds.Url,
		s.Rds.Port,
		instance.Database)

	credentials := map[string]string{
		"uri":      uri,
		"username": instance.Username,
		"password": password,
		"host":     s.Rds.Url,
		"db_name":  instance.Database,
	}

	response := map[string]interface{}{
		"credentials": credentials,
	}
	r.JSON(201, response)
}

// DeleteInstance
// URL: /v2/service_instances/:id
// Params:
// {
//   "service_id": "service-id-here"
//   "plan_id":    "plan-id-here"
// }
func DeleteInstance(p martini.Params, r render.Render, db *gorm.DB) {
	instance := Instance{}

	db.Where("uuid = ?", p["id"]).First(&instance)

	if instance.Id == 0 {
		r.JSON(404, Response{"Instance not found"})
		return
	}

	db.Exec(fmt.Sprintf("DROP DATABASE %s;", instance.Database))
	db.Exec(fmt.Sprintf("DROP USER %s;", instance.Username))

	db.Delete(&instance)

	r.JSON(200, Response{"The instance was deleted"})
}
