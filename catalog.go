package main

type Metadata struct {
	DisplayName         string `json:"displayName"`
	ImageUrl            string `json:"imageUrl"`
	LongDescription     string `json:"longDescription"`
	ProviderDisplayName string `json:"providerDisplayName"`
	DocumentationUrl    string `json:"documentationUrl"`
	SupportUrl          string `json:"supportUrl"`
}
type PlanCost struct {
	Amount map[string]float64 `json:"amount"`
	Unit   string             `json:"unit"`
}
type PlanMetadata struct {
	Bullets     []string   `json:"bullets"`
	Costs       []PlanCost `json:"costs"`
	DisplayName string     `json:"displayName"`
}
type Plan struct {
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Metadata     PlanMetadata `json:"metadata"`
	Free         bool         `json:"free"`
	Adapter      string       `json:"-"`
	InstanceType string       `json:"-"`
	DbType       string       `json:"-"`
}

type Service struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Bindable    bool     `json:"bindable"`
	Tags        []string `json:"tags"`
	Metadata    Metadata `json:"metadata"`
	Plans       []Plan   `json:"plans"`
}

func BuildCatalog() []Service {

	service := Service{
		Id:          "db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
		Name:        "rds",
		Description: "RDS Database Broker",
		Bindable:    true,
		Tags:        []string{"database", "RDS", "postgresql"},
		Metadata: Metadata{
			DisplayName:         "RDS Database Broker",
			ProviderDisplayName: "RDS",
		},
		Plans: GetPlans(),
	}

	return []Service{service}
}

func GetPlans() []Plan {
	sharedPlan := Plan{
		Id:          "44d24fc7-f7a4-4ac1-b7a0-de82836e89a3",
		Name:        "shared-psql",
		Description: "Shared infrastructure for Postgres DB",
		Metadata: PlanMetadata{
			Bullets: []string{"Shared RDS Instance", "Postgres instance"},
			Costs: []PlanCost{
				PlanCost{
					Amount: map[string]float64{
						"usd": 0.00,
					},
					Unit: "MONTHLY",
				},
			},
			DisplayName: "Free Shared Plan",
		},
		Free:    true,
		Adapter: "shared",
		DbType:  "postgres",
	}

	microPlan := Plan{
		Id:          "da91e15c-98c9-46a9-b114-02b8d28062c6",
		Name:        "micro-psql",
		Description: "Dedicated Micro RDS Postgres DB Instance",
		Metadata: PlanMetadata{
			Bullets: []string{"Dedicated Redundant RDS Instance", "Postgres instance"},
			Costs: []PlanCost{
				PlanCost{
					Amount: map[string]float64{
						"usd": 0.036,
					},
					Unit: "HOURLY",
				},
			},
			DisplayName: "Dedicated Micro Postgres",
		},
		Free:         false,
		Adapter:      "dedicated",
		InstanceType: "db.t2.micro",
		DbType:  "postgres",
	}

	mediumPlan := Plan{
		Id:          "332e0168-6969-4bd7-b07f-29f08c4bf78e",
		Name:        "medium-psql",
		Description: "Dedicated Medium RDS Postgres DB Instance",
		Metadata: PlanMetadata{
			Bullets: []string{"Dedicated Redundant RDS Instance", "Postgres instance"},
			Costs: []PlanCost{
				PlanCost{
					Amount: map[string]float64{
						"usd": 0.190,
					},
					Unit: "HOURLY",
				},
			},
			DisplayName: "Dedicated Medium Postgres",
		},
		Free:         false,
		Adapter:      "dedicated",
		InstanceType: "db.m3.medium",
		DbType:  "postgres",
	}

	return []Plan{sharedPlan, microPlan, mediumPlan}
}

func FindPlan(id string) *Plan {
	for _, p := range GetPlans() {
		if p.Id == id {
			return &p
		}
	}

	return nil
}
