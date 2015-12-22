package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ServiceMetadata struct {
	DisplayName         string `yaml:"displayName" json:"displayName"`
	ImageUrl            string `yaml:"imageUrl" json:"imageUrl"`
	LongDescription     string `yaml:"longDescription" json:"longDescription"`
	ProviderDisplayName string `yaml:"providerDisplayName" json:"providerDisplayName"`
	DocumentationUrl    string `yaml:"documentationUrl" json:"documentationUrl"`
	SupportUrl          string `yaml:"supportUrl" json:"supportUrl"`
}

type PlanCost struct {
	Amount map[string]float64 `yaml:"amount" json:"amount"`
	Unit   string             `yaml:"unit" json:"unit"`
}

type PlanMetadata struct {
	Bullets     []string   `yaml:"bullets" json:"bullets"`
	Costs       []PlanCost `yaml:"costs" json:"costs"`
	DisplayName string     `yaml:"displayName" json:"displayName"`
}

type Plan struct {
	Id           string       `yaml:"id" json:"id"`
	Name         string       `yaml:"name" json:"name"`
	Description  string       `yaml:"description" json:"description"`
	Metadata     PlanMetadata `yaml:"metadata" json:"metadata"`
	Free         bool         `yaml:"free" json:"free"`
	Adapter      string       `yaml:"adapter" json:"-"`
	InstanceType string       `yaml:"instanceType" json:"-"`
	DbType       string       `yaml:"dbType" json:"-"`
}

type Service struct {
	Id          string          `yaml:"id" json:"id"`
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description" json:"description"`
	Bindable    bool            `yaml:"bindable" json:"bindable"`
	Tags        []string        `yaml:"tags" json:"tags"`
	Metadata    ServiceMetadata `yaml:"metadata" json:"metadata"`
	Plans       []Plan          `yaml:"plans" json:"plans"`
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
		if service.Id == serviceID {
			return &service
		}
	}
	return nil
}

// fetchPlan return a pointer to a Plan struct with the given plan ID
func (catalog *Catalog) fetchPlan(serviceID string, planID string) *Plan {
	service := catalog.fetchService(serviceID)
	if service != nil {
		for _, plan := range service.Plans {
			if plan.Id == planID {
				return &plan
			}
		}
	}
	return nil
}
