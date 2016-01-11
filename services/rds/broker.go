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
func initializeAdapter(plan catalog.RDSPlan, s *config.Settings, c *catalog.Catalog) (DBAdapter, response.Response) {

	var dbAdapter DBAdapter
	// For test environments, use a mock adapter.
	if s.Environment == "test" {
		dbAdapter = &MockDBAdapter{}
		return dbAdapter, nil
	}

	switch plan.Adapter {
	case "shared":
		setting, err := c.GetResources().RdsSettings.GetRDSSettingByPlan(plan.ID)
		if err != nil {
			return nil, response.NewErrorResponse(http.StatusInternalServerError, err.Error())
		}
		if setting.DB == nil {
			return nil, response.NewErrorResponse(http.StatusInternalServerError, "An internal error occurred setting up shared databases.")
		}
		dbAdapter = &SharedDBAdapter{
			SharedDbConn: setting.DB,
		}
	case "dedicated":
		dbAdapter = &DedicatedDBAdapter{
			InstanceClass: plan.InstanceClass,
		}
	default:
		return nil, response.NewErrorResponse(http.StatusInternalServerError, "Adapter not found")
	}

	return dbAdapter, nil
}

func InitRDSBroker(brokerDB *gorm.DB, settings *config.Settings) base.Broker {
	return &rdsBroker{brokerDB, settings}
}

func (broker *rdsBroker) CreateInstance(c *catalog.Catalog, id string, createRequest request.Request) response.Response {
	newInstance := RDSInstance{}

	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&newInstance).Count(&count)
	if count != 0 {
		return response.NewErrorResponse(http.StatusConflict, "The instance already exists")
	}

	plan, planErr := c.RdsService.FetchPlan(createRequest.PlanId)
	if planErr != nil {
		return planErr
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

	adapter, adapterErr := initializeAdapter(plan, broker.settings, c)
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

	if newInstance.Adapter == "shared" {
		setting, err := c.GetResources().RdsSettings.GetRDSSettingByPlan(plan.ID)
		if err != nil {
			return response.NewErrorResponse(http.StatusInternalServerError, err.Error())
		}
		newInstance.Host = setting.Config.Url
		newInstance.Port = setting.Config.Port
	}
	err = broker.brokerDB.Save(&newInstance).Error
	if err != nil {
		return response.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	return response.NewSuccessCreateResponse()
}

func (broker *rdsBroker) BindInstance(c *catalog.Catalog, id string, baseInstance base.Instance) response.Response {
	existingInstance := RDSInstance{}

	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&existingInstance).Count(&count)
	if count == 0 {
		return response.NewErrorResponse(http.StatusNotFound, "Instance not found")
	}

	plan, planErr := c.RdsService.FetchPlan(baseInstance.PlanId)
	if planErr != nil {
		return planErr
	}

	password, err := existingInstance.GetPassword(broker.settings.EncryptionKey)
	if err != nil {
		return response.NewErrorResponse(http.StatusInternalServerError, "Unable to get instance password.")
	}

	// Get the correct database logic depending on the type of plan. (shared vs dedicated)
	adapter, adapterErr := initializeAdapter(plan, broker.settings, c)
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

func (broker *rdsBroker) DeleteInstance(c *catalog.Catalog, id string, baseInstance base.Instance) response.Response {
	existingInstance := RDSInstance{}
	var count int64
	broker.brokerDB.Where("uuid = ?", id).First(&existingInstance).Count(&count)
	if count == 0 {
		return response.NewErrorResponse(http.StatusNotFound, "Instance not found")
	}

	plan, planErr := c.RdsService.FetchPlan(baseInstance.PlanId)
	if planErr != nil {
		return planErr
	}

	adapter, adapterErr := initializeAdapter(plan, broker.settings, c)
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
