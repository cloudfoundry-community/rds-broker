package catalog

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"errors"
	"fmt"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"gopkg.in/go-playground/validator.v8"
	"gopkg.in/yaml.v2"
	"net/http"
	"reflect"
)

// ServiceMetadata contains the service metadata fields listed in the Cloud Foundry docs:
// http://docs.cloudfoundry.org/services/catalog-metadata.html#services-metadata-fields
type ServiceMetadata struct {
	DisplayName         string `yaml:"displayName" json:"displayName"`
	ImageURL            string `yaml:"imageUrl" json:"imageUrl"`
	LongDescription     string `yaml:"longDescription" json:"longDescription"`
	ProviderDisplayName string `yaml:"providerDisplayName" json:"providerDisplayName"`
	DocumentationURL    string `yaml:"documentationUrl" json:"documentationUrl"`
	SupportURL          string `yaml:"supportUrl" json:"supportUrl"`
}

// PlanCost contains an array-of-objects that describes the costs of a service,
// in what currency, and the unit of measure.
type PlanCost struct {
	Amount map[string]float64 `yaml:"amount" json:"amount" validate:"required"`
	Unit   string             `yaml:"unit" json:"unit" validate:"required"`
}

// PlanMetadata contains the plan metadata fields listed in the Cloud Foundry docs:
// http://docs.cloudfoundry.org/services/catalog-metadata.html#plan-metadata-fields
type PlanMetadata struct {
	Bullets     []string   `yaml:"bullets" json:"bullets"`
	Costs       []PlanCost `yaml:"costs" json:"costs"`
	DisplayName string     `yaml:"displayName" json:"displayName"`
}

// Plan is a generic struct for a Cloud Foundry service plan
// http://docs.cloudfoundry.org/services/api.html
type Plan struct {
	ID          string       `yaml:"id" json:"id" validate:"required"`
	Name        string       `yaml:"name" json:"name" validate:"required"`
	Description string       `yaml:"description" json:"description" validate:"required"`
	Metadata    PlanMetadata `yaml:"metadata" json:"metadata" validate:"required"`
	Free        bool         `yaml:"free" json:"free"`
}

var (
	ErrNoServiceFound = errors.New("No service found for given service id.")
	ErrNoPlanFound    = errors.New("No plan found for given plan id.")
)

type RDSService struct {
	Service `yaml:",inline" validate:"required"`
	Plans   []RDSPlan `yaml:"plans" json:"plans" validate:"required,dive,required"`
}

// RDSPlan inherits from a Plan and adds fields specific to AWS.
// these fields are read from the catalog.yaml file, but are not rendered
// in the catalog API endpoint.
type RDSPlan struct {
	Plan          `yaml:",inline" validate:"required"`
	Adapter       string `yaml:"adapter" json:"-" validate:"required"`
	InstanceClass string `yaml:"instanceClass" json:"-"`
	DbType        string `yaml:"dbType" json:"-" validate:"required"`
}

func (s RDSService) FetchPlan(planId string) (RDSPlan, response.Response) {
	for _, plan := range s.Plans {
		if plan.ID == planId {
			return plan, nil
		}
	}
	return RDSPlan{}, response.NewErrorResponse(http.StatusBadRequest, ErrNoPlanFound.Error())
}

// Catalog struct holds a collections of services
type Catalog struct {
	RdsService RDSService `yaml:"rds" json:"-"`
}

// Service struct contains data for the Cloud Foundry service
// http://docs.cloudfoundry.org/services/api.html
type Service struct {
	ID          string          `yaml:"id" json:"id" validate:"required"`
	Name        string          `yaml:"name" json:"name" validate:"required"`
	Description string          `yaml:"description" json:"description" validate:"required"`
	Bindable    bool            `yaml:"bindable" json:"bindable" validate:"required"`
	Tags        []string        `yaml:"tags" json:"tags" validate:"required"`
	Metadata    ServiceMetadata `yaml:"metadata" json:"metadata" validate:"required"`
}

func (c *Catalog) GetServices() []interface{} {
	catalogStruct := reflect.ValueOf(*c)
	numOfFields := catalogStruct.NumField()
	services := make([]interface{}, numOfFields)
	for i := 0; i < numOfFields; i++ {
		services[i] = catalogStruct.Field(i).Interface()
	}
	return services
}

// InitCatalog initalizes a Catalog struct that contains services and plans
// defined in the catalog.yaml configuation file and returns a pointer to that catalog
func InitCatalog(path string) *Catalog {
	var catalog Catalog
	catalogFile := filepath.Join(path, "catalog.yaml")
	data, err := ioutil.ReadFile(catalogFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(data, &catalog)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config := &validator.Config{TagName: "validate"}

	validate := validator.New(config)
	validateErr := validate.Struct(catalog)
	if validateErr != nil {
		fmt.Println(validateErr)
		return nil
	}
	fmt.Printf("%+v\n", catalog)
	return &catalog
}
