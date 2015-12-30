package main

import (
	"github.com/cloudfoundry-community/aws-broker/base"
	"github.com/cloudfoundry-community/aws-broker/catalog"
	"strings"
	"github.com/jinzhu/gorm"
	"github.com/cloudfoundry-community/aws-broker/services/rds"
	"encoding/json"
	//"errors"
	"net/http"
	"github.com/cloudfoundry-community/aws-broker/helpers"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"io/ioutil"
	"github.com/cloudfoundry-community/aws-broker/config"
)

var (
	ErrNoRequestResponse = response.New(http.StatusBadRequest, "No Request Body")
)

func extractServiceReq(req *http.Request) (helpers.ServiceReq, *response.Response) {
	var sr helpers.ServiceReq
	if req.Body == nil {
		return sr, ErrNoRequestResponse
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return sr, response.New(http.StatusBadRequest, err.Error())
	}
	json.Unmarshal(body, &sr)
	return sr, nil
}

func findBroker(serviceReq helpers.ServiceReq, c *catalog.Catalog, brokerDb *gorm.DB, settings *config.Settings) (base.Broker, *response.Response) {
	// Look in catalog and find the service.
	service, err := c.FetchService(serviceReq.ServiceId)
	if err != nil {
		return nil, response.New(http.StatusNotFound, err.Error())
	}
	switch strings.ToLower(service.Name) {
	case "rds":
		return rds.InitRDSBroker(serviceReq, brokerDb, settings), nil
	}
	return nil, nil
}

func createInstance(req *http.Request, c *catalog.Catalog, brokerDb *gorm.DB, id string, settings *config.Settings) *response.Response {
	serviceReq, resp := extractServiceReq(req)
	if resp != nil {
		return resp
	}
	broker, resp := findBroker(serviceReq, c, brokerDb, settings)
	if resp != nil {
		return resp
	}

	plan, planErr := c.FetchPlan(serviceReq.ServiceId, serviceReq.PlanId)
	if planErr != nil {
		return response.New(http.StatusBadRequest, planErr.Error())
	}

	// Create instance
	return broker.CreateInstance(plan, id)
}

func bindInstance(broker base.Broker, serviceReq helpers.ServiceReq) *response.Response {
	// Get the existing instance.

	// Bind it to the app.
	return nil
}

func deleteInstance(broker base.Broker, serviceReq helpers.ServiceReq) *response.Response {
	// Get the existing instance.

	// Delete it and update the broker database.
	return nil
}
