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
)

func findBroker(serviceId string, c *catalog.Catalog, brokerDb *gorm.DB, settings *config.Settings) (base.Broker, response.Response) {
	switch serviceId {
	// RDS Service
	case c.RdsService.ID:
		return rds.InitRDSBroker(brokerDb, settings), nil
	}

	return nil, response.NewErrorResponse(http.StatusNotFound, catalog.ErrNoServiceFound.Error())
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

	// Create instance
	return broker.CreateInstance(c, id, createRequest)
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

	return broker.BindInstance(c, id, bindRequest)
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

	return broker.DeleteInstance(c, id, deleteRequest)
}
