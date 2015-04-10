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
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Metadata    PlanMetadata `json:"metadata"`
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
	freePlan := Plan{
		Id:          "44d24fc7-f7a4-4ac1-b7a0-de82836e89a3",
		Name:        "shared",
		Description: "Shared infrastructure for DB",
		Metadata: PlanMetadata{
			Bullets: []string{"Shared RDS Instance"},
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
	}
	service := Service{
		Id:          "db80ca29-2d1b-4fbc-aad3-d03c0bfa7593",
		Name:        "rds-database",
		Description: "RDS Database Broker",
		Bindable:    true,
		Tags:        []string{"database", "RDS", "postgresql"},
		Metadata: Metadata{
			DisplayName:         "RDS Database Broker",
			ProviderDisplayName: "RDS",
		},
		Plans: []Plan{freePlan},
	}

	return []Service{service}
}
