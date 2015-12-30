package main

import (
	"github.com/cloudfoundry-community/aws-broker/config"
	// "github.com/cloudfoundry-community/aws-broker/base"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	//"encoding/json"
	//"io/ioutil"
	"net/http"
	"github.com/cloudfoundry-community/aws-broker/catalog"
	//"errors"
)


/*
type Operation struct {
	State                    string
	Description              string
	AsyncPollIntervalSeconds int `json:"async_poll_interval_seconds, omitempty"`
}

type CreateResponse struct {
	DashboardUrl  string
	LastOperation Operation
}
*/


// CreateInstance
// URL: /v2/service_instances/:id
// Request:
// {
//   "service_id":        "service-guid-here",
//   "plan_id":           "plan-guid-here",
//   "organization_guid": "org-guid-here",
//   "space_guid":        "space-guid-here"
// }
func CreateInstance(p martini.Params, req *http.Request, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	resp := createInstance(req, c, brokerDb, p["id"], s)
	r.JSON(resp.StatusCode, resp)
}

// BindInstance
// URL: /v2/service_instances/:instance_id/service_bindings/:binding_id
// Request:
// {
//   "plan_id":        "plan-guid-here",
//   "service_id":     "service-guid-here",
//   "app_guid":       "app-guid-here"
// }
func BindInstance(p martini.Params, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	/*
	existingInstance := base.Instance{}

	brokerDb.Where("uuid = ?", p["instance_id"]).First(&existingInstance)
	if existingInstance.Id == 0 {
		r.JSON(404, Response{"Instance not found"})
		return
	}
	password, err := existingInstance.GetPassword(s.EncryptionKey)
	if err != nil {
		r.JSON(http.StatusInternalServerError, "Unable to get instance password.")
	}

	plan, planErr := c.FetchPlan(existingInstance.ServiceId, existingInstance.PlanId)

	if planErr != nil {
		r.JSON(http.StatusBadRequest, Response{planErr.Error()})
		return
	}

	// Get the correct database logic depending on the type of plan. (shared vs dedicated)
	db, err := s.InitializeAdapter(plan, brokerDb)
	if err != nil {
		desc := "There was an error creating the instance. Error: " + err.Error()
		r.JSON(http.StatusInternalServerError, Response{desc})
		return
	}

	var credentials map[string]string
	// Bind the database instance to the application.
	originalInstanceState := existingInstance.State
	if credentials, err = db.BindDBToApp(&existingInstance, password); err != nil {
		desc := "There was an error binding the database instance to the application."
		if err != nil {
			desc = desc + " Error: " + err.Error()
		}
		r.JSON(http.StatusInternalServerError, Response{desc})
		return
	}

	// If the state of the instance has changed, update it.
	if existingInstance.State != originalInstanceState {
		brokerDb.Save(&existingInstance)
	}

	response := map[string]interface{}{
		"credentials": credentials,
	}
	r.JSON(http.StatusCreated, response)
	*/
}

// DeleteInstance
// URL: /v2/service_instances/:id
// Request:
// {
//   "service_id": "service-id-here"
//   "plan_id":    "plan-id-here"
// }
func DeleteInstance(p martini.Params, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	/*
	existingInstance := base.Instance{}

	brokerDb.Where("uuid = ?", p["id"]).First(&existingInstance)

	if existingInstance.Id == 0 {
		r.JSON(http.StatusNotFound, Response{"Instance not found"})
		return
	}

	plan, planErr := c.FetchPlan(existingInstance.ServiceId, existingInstance.PlanId)

	if planErr != nil {
		r.JSON(http.StatusBadRequest, Response{planErr.Error()})
		return
	}
	// Get the correct database logic depending on the type of plan. (shared vs dedicated)
	db, err := s.InitializeAdapter(plan, brokerDb)
	if err != nil {
		desc := "There was an error deleting the instance. Error: " + err.Error()
		r.JSON(http.StatusInternalServerError, Response{desc})
		return
	}
	var status base.InstanceState
	// Delete the database instance.
	if status, err = db.DeleteDB(&existingInstance); status == base.InstanceNotGone {
		desc := "There was an error deleting the instance."
		if err != nil {
			desc = desc + " Error: " + err.Error()
		}
		r.JSON(http.StatusInternalServerError, Response{desc})
		return
	}
	brokerDb.Delete(&existingInstance)
	r.JSON(http.StatusOK, Response{"The instance was deleted"})
	*/
}
