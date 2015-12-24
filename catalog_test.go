package main

import "testing"

// Create a catalog
var catalog = initCatalog()

func TestInitCatalog(t *testing.T) {
	if len(catalog.Services) == 0 {
		t.Error("There should be at least 1 service")
	}
}

func TestFetchPlan(t *testing.T) {
	plan := catalog.fetchPlan(
		"db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
		"44d24fc7-f7a4-4ac1-b7a0-de82836e89a3",
	)

	if plan == nil {
		t.Error("Plan has to exist")
	}

	if plan.Name != "shared-psql" {
		t.Error("Found the wrong plan")
	}
}

func TestFetchService(t *testing.T) {
	service := catalog.fetchService(
		"db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
	)

	if service == nil {
		t.Error("Service has to exist")
	}

	if service.Name != "rds" {
		t.Error("Found the wrong service")
	}
}
