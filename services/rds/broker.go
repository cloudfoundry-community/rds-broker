package rds

import (
	"github.com/cloudfoundry-community/aws-broker/helpers"
	"github.com/jinzhu/gorm"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/cloudfoundry-community/aws-broker/catalog"
	"github.com/cloudfoundry-community/aws-broker/config"
	"net/http"
	"github.com/cloudfoundry-community/aws-broker/base"
)

type rdsBroker struct {
	brokerDB *gorm.DB
	serviceReq helpers.ServiceReq
	settings *config.Settings
}


// InitializeAdapter is the main function to create database instances
func initializeAdapter(plan catalog.AWSPlan, s *config.Settings,
sharedDbConn *gorm.DB) (DBAdapter, *response.Response) {

	var dbAdapter DBAdapter
	// For test environments, use a mock adapter.
	if s.Environment == "test" {
		dbAdapter = &MockDBAdapter{}
		return dbAdapter, nil
	}

	switch plan.Adapter {
	case "shared":
		dbAdapter = &SharedDBAdapter{
			SharedDbConn: sharedDbConn,
		}
	case "dedicated":
		dbAdapter = &DedicatedDBAdapter{
			InstanceType: plan.InstanceType,
		}
	default:
		return nil, response.New(http.StatusInternalServerError, "Adapter not found")
	}

	return dbAdapter, nil
}

func InitRDSBroker(serviceReq helpers.ServiceReq, brokerDB *gorm.DB, settings *config.Settings) base.Broker {
	return &rdsBroker{brokerDB, serviceReq, settings}
}

func (broker *rdsBroker) CreateInstance(plan catalog.AWSPlan, id string) *response.Response {
	newInstance := RDSInstance{}

	broker.brokerDB.Where("uuid = ?", id).First(&newInstance)

	if newInstance.Id > 0 {
		return response.New(http.StatusConflict, "The instance already exists")
	}
	// TODO Check for exisiting.
	err := newInstance.Init(
		id,
		broker.serviceReq.OrganizationGuid,
		broker.serviceReq.SpaceGuid,
		broker.serviceReq.ServiceId,
		plan,
		broker.settings)

	if err != nil {
		return response.New(http.StatusBadRequest, "There was an error initializing the instance. Error: " + err.Error())
	}

	adapter, adapterErr := initializeAdapter(plan, broker.settings, broker.brokerDB)
	if adapterErr != nil {
		return adapterErr
	}
	// Create the database instance.
	status, err := adapter.CreateDB(&newInstance, newInstance.ClearPassword)
	if status == base.InstanceNotCreated {
		desc := "There was an error creating the instance."
		if err != nil {
			desc = desc + " Error: " + err.Error()
		}
		return response.New(http.StatusBadRequest, desc)
	}

	newInstance.State = status

	// FIXME
	// Currently, if we are dealing with a shared database, it will not populate the host and port fields of the instance.
	// Also, currently, the shared database instance just create a new database and user inside the internal broker database.
	// Eventually we want to register a DBConfig or a pool of database connections for the shared instances to get the host and port
	// and move the logic of storing it in the instance in the SharedDB's CreateDB.
	if newInstance.Adapter == "shared" {
		newInstance.Host = broker.settings.DbConfig.Url
		newInstance.Port = broker.settings.DbConfig.Port
	}
	broker.brokerDB.Save(&newInstance)
	return response.New(http.StatusCreated, "The instance was created")
}

func (broker *rdsBroker) BindInstance(plan catalog.AWSPlan, id string) *response.Response {
	existingInstance := RDSInstance{}

	broker.brokerDB.Where("uuid = ?", id).First(&existingInstance)
	if existingInstance.Id == 0 {
		return response.New(http.StatusNotFound, "Instance not found")
	}
	password, err := existingInstance.GetPassword(broker.settings.EncryptionKey)
	if err != nil {
		return response.New(http.StatusInternalServerError, "Unable to get instance password.")
	}

	//plan, planErr := c.FetchPlan(existingInstance.ServiceId, existingInstance.PlanId)

	/*
	if planErr != nil {
		r.JSON(http.StatusBadRequest, Response{planErr.Error()})
		return
	}
	*/

	// Get the correct database logic depending on the type of plan. (shared vs dedicated)
	adapter, adapterErr := initializeAdapter(plan, broker.settings, broker.brokerDB)
	if adapterErr != nil {
		return adapterErr
	}

	var credentials map[string]string
	// Bind the database instance to the application.
	originalInstanceState := existingInstance.State
	if credentials, err = adapter.BindDBToApp(&existingInstance, password); err != nil {
		desc := "There was an error binding the database instance to the application."
		if err != nil {
			desc = desc + " Error: " + err.Error()
		}
		return response.New(http.StatusBadRequest, desc)
	}

	// If the state of the instance has changed, update it.
	if existingInstance.State != originalInstanceState {
		broker.brokerDB.Save(&existingInstance)
	}

	_ = map[string]interface{}{
		"credentials": credentials,
	}
	//return response.New(http.StatusCreated, "The instance was created")
	//r.JSON(http.StatusCreated, response)
	return nil
}

func (broker *rdsBroker) UnbindInstance() *response.Response {
	return nil
}

func (broker *rdsBroker) DeleteInstance() *response.Response {
	return nil
}
