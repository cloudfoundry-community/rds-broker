package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
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
	Amount map[string]float64 `yaml:"amount" json:"amount"`
	Unit   string             `yaml:"unit" json:"unit"`
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
	ID          string       `yaml:"id" json:"id"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Metadata    PlanMetadata `yaml:"metadata" json:"metadata"`
	Free        bool         `yaml:"free" json:"free"`
}

// AWSPlan inherits from a Plan and adds fields specific to AWS.
// these fields are read from the catalog.yaml file, but are not rendered
// in the catalog API endpoint.
type AWSPlan struct {
	Plan         `yaml:",inline"`
	Adapter      string `yaml:"adapter" json:"-"`
	InstanceType string `yaml:"instanceType" json:"-"`
	DbType       string `yaml:"dbType" json:"-"`
}

// Service struct contains data for the Cloud Foundry service
// http://docs.cloudfoundry.org/services/api.html
type Service struct {
	ID          string          `yaml:"id" json:"id"`
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description" json:"description"`
	Bindable    bool            `yaml:"bindable" json:"bindable"`
	Tags        []string        `yaml:"tags" json:"tags"`
	Metadata    ServiceMetadata `yaml:"metadata" json:"metadata"`
	Plans       []AWSPlan       `yaml:"plans" json:"plans"`
}

// Catalog struct holds a collections of services
type Catalog struct {
	Services []Service `yaml:"services" json:"services"`
}

// initCatalog initalizes a Catalog struct that contains services and plans
// defined in the catalog.yaml configuation file and returns a pointer to that catalog
func initCatalog() *Catalog {
	var catalog Catalog
	workingDir, _ := os.Getwd()
	catalogFile := filepath.Join(workingDir, "catalog.yaml")
	data, err := ioutil.ReadFile(catalogFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(data, &catalog)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &catalog
}

// fetchService returns a pointer to the Service struct with the given service ID
func (catalog *Catalog) fetchService(serviceID string) *Service {
	for _, service := range catalog.Services {
		if service.ID == serviceID {
			return &service
		}
	}
	return nil
}

// fetchPlan return a pointer to a Plan struct with the given plan ID
func (catalog *Catalog) fetchPlan(serviceID string, planID string) *AWSPlan {
	service := catalog.fetchService(serviceID)
	if service != nil {
		for _, plan := range service.Plans {
			if plan.ID == planID {
				return &plan
			}
		}
	}
	return nil
}
