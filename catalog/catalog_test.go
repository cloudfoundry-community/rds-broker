package catalog

import "testing"

// Create a catalog
var catalog = InitCatalog()

func TestInitCatalog(t *testing.T) {
	if len(catalog.Services) == 0 {
		t.Error("There should be at least 1 service")
	}
}

func TestFetchPlan(t *testing.T) {
	plan, err := catalog.FetchPlan(
		"db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
		"44d24fc7-f7a4-4ac1-b7a0-de82836e89a3",
	)

	if err != nil {
		t.Error("Plan has to exist")
	}

	if plan.Name != "shared-psql" {
		t.Error("Found the wrong plan")
	}
}

func TestFetchService(t *testing.T) {
	service, err := catalog.FetchService(
		"db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
	)

	if err != nil {
		t.Error("Service has to exist")
	}

	if service.Name != "rds" {
		t.Error("Found the wrong service")
	}
}
