package main

import (
	"github.com/cloudfoundry-community/aws-broker/base"
	"github.com/cloudfoundry-community/aws-broker/catalog"
	"github.com/cloudfoundry-community/aws-broker/config"
	"github.com/cloudfoundry-community/aws-broker/helpers/request"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/cloudfoundry-community/aws-broker/services/rds"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"
)

func findBroker(serviceId string, c *catalog.Catalog, brokerDb *gorm.DB, settings *config.Settings) (base.Broker, response.Response) {
	// Look in catalog and find the service.
	service, err := c.FetchService(serviceId)
	if err != nil {
		return nil, response.NewErrorResponse(http.StatusNotFound, err.Error())
	}
	switch strings.ToLower(service.Name) {
	case "rds":
		return rds.InitRDSBroker(brokerDb, settings), nil
	}
	return nil, nil
}

func createInstance(req *http.Request, c *catalog.Catalog, brokerDb *gorm.DB, id string, settings *config.Settings) response.Response {
	createRequest, resp := request.ExtractCreateRequest(req)
	if resp != nil {
		return resp
	}
	broker, resp := findBroker(createRequest.ServiceId, c, brokerDb, settings)
	if resp != nil {
		return resp
	}

	plan, planErr := c.FetchPlan(createRequest.ServiceId, createRequest.PlanId)
	if planErr != nil {
		return response.NewErrorResponse(http.StatusBadRequest, planErr.Error())
	}

	// Create instance
	return broker.CreateInstance(plan, id, createRequest)
}

func bindInstance(req *http.Request, c *catalog.Catalog, brokerDb *gorm.DB, id string, settings *config.Settings) response.Response {
	bindRequest, resp := request.ExtractBindRequest(req)
	if resp != nil {
		return resp
	}
	broker, resp := findBroker(bindRequest.ServiceId, c, brokerDb, settings)
	if resp != nil {
		return resp
	}

	plan, planErr := c.FetchPlan(bindRequest.ServiceId, bindRequest.PlanId)
	if planErr != nil {
		return response.NewErrorResponse(http.StatusBadRequest, planErr.Error())
	}

	return broker.BindInstance(plan, id)
}

func deleteInstance(req *http.Request, c *catalog.Catalog, brokerDb *gorm.DB, id string, settings *config.Settings) response.Response {
	deleteRequest, resp := request.ExtractDeleteRequest(req)
	if resp != nil {
		return resp
	}
	broker, resp := findBroker(deleteRequest.ServiceId, c, brokerDb, settings)
	if resp != nil {
		return resp
	}

	plan, planErr := c.FetchPlan(deleteRequest.ServiceId, deleteRequest.PlanId)
	if planErr != nil {
		return response.NewErrorResponse(http.StatusBadRequest, planErr.Error())
	}

	return broker.DeleteInstance(plan, id)
}
