package main

import (
	"testing"
)

func TestGetPlans(t *testing.T) {
	plans := GetPlans()

	if len(plans) == 0 {
		t.Error("There should be at least 1 plan")
	}
}

func TestFindPlan(t *testing.T) {
	plan := FindPlan("44d24fc7-f7a4-4ac1-b7a0-de82836e89a3")

	if plan == nil {
		t.Error("Plan has to exist")
	}

	if plan.Name != "shared-psql" {
		t.Error("Found the wrong plan")
	}
}
