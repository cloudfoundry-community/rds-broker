package rds

import (
	"github.com/cloudfoundry-community/aws-broker/base"
	"github.com/cloudfoundry-community/aws-broker/catalog"
	"github.com/cloudfoundry-community/aws-broker/config"
	"github.com/cloudfoundry-community/aws-broker/helpers/request"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/jinzhu/gorm"
	"net/http"
)

type rdsBroker struct {
	brokerDB *gorm.DB
	settings *config.Settings
}

// InitializeAdapter is the main function to create database instances
func initializeAdapter(plan catalog.AWSPlan, s *config.Settings,
	sharedDbConn *gorm.DB) (DBAdapter, response.Response) {

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
		return nil, response.NewErrorResponse(http.StatusInternalServerError, "Adapter not found")
	}

	return dbAdapter, nil
}

func InitRDSBroker(brokerDB *gorm.DB, settings *config.Settings) base.Broker {
	return &rdsBroker{brokerDB, settings}
}

func (broker *rdsBroker) CreateInstance(plan catalog.AWSPlan, id string, createRequest request.CreateRequest) response.Response {
	newInstance := RDSInstance{}

	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&newInstance).Count(&count)
	if count != 0 {
		return response.NewErrorResponse(http.StatusConflict, "The instance already exists")
	}

	err := newInstance.Init(
		id,
		createRequest.OrganizationGuid,
		createRequest.SpaceGuid,
		createRequest.ServiceId,
		plan,
		broker.settings)

	if err != nil {
		return response.NewErrorResponse(http.StatusBadRequest, "There was an error initializing the instance. Error: "+err.Error())
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
		return response.NewErrorResponse(http.StatusBadRequest, desc)
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
	return response.NewSuccessCreateResponse()
}

func (broker *rdsBroker) BindInstance(plan catalog.AWSPlan, id string) response.Response {
	existingInstance := RDSInstance{}

	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&existingInstance).Count(&count)
	if count == 0 {
		return response.NewErrorResponse(http.StatusNotFound, "Instance not found")
	}

	password, err := existingInstance.GetPassword(broker.settings.EncryptionKey)
	if err != nil {
		return response.NewErrorResponse(http.StatusInternalServerError, "Unable to get instance password.")
	}

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
		return response.NewErrorResponse(http.StatusBadRequest, desc)
	}

	// If the state of the instance has changed, update it.
	if existingInstance.State != originalInstanceState {
		broker.brokerDB.Save(&existingInstance)
	}

	return response.NewSuccessBindResponse(credentials)
}

func (broker *rdsBroker) DeleteInstance(plan catalog.AWSPlan, id string) response.Response {
	existingInstance := RDSInstance{}
	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&existingInstance).Count(&count)
	if count == 0 {
		return response.NewErrorResponse(http.StatusNotFound, "Instance not found")
	}

	adapter, adapterErr := initializeAdapter(plan, broker.settings, broker.brokerDB)
	if adapterErr != nil {
		return adapterErr
	}
	// Delete the database instance.
	if status, err := adapter.DeleteDB(&existingInstance); status == base.InstanceNotGone {
		desc := "There was an error deleting the instance."
		if err != nil {
			desc = desc + " Error: " + err.Error()
		}
		return response.NewErrorResponse(http.StatusBadRequest, desc)
	}
	broker.brokerDB.Delete(&existingInstance)
	return response.NewSuccessDeleteResponse()
}
