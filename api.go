package main

import (
	"github.com/18F/aws-broker/config"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	"github.com/18F/aws-broker/catalog"
	"net/http"
)

/*
type Operation struct {
	State                    string
	Description              string
	AsyncPollIntervalSeconds int `json:"async_poll_interval_seconds, omitempty"`
}

type CreateResponse struct {
	DashboardUrl  string
	LastOperation Operation
}
*/

// CreateInstance
// URL: /v2/service_instances/:id
// Request:
// {
//   "service_id":        "service-guid-here",
//   "plan_id":           "plan-guid-here",
//   "organization_guid": "org-guid-here",
//   "space_guid":        "space-guid-here"
// }
func CreateInstance(p martini.Params, req *http.Request, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	resp := createInstance(req, c, brokerDb, p["id"], s)
	r.JSON(resp.GetStatusCode(), resp)
}

// BindInstance
// URL: /v2/service_instances/:instance_id/service_bindings/:binding_id
// Request:
// {
//   "plan_id":        "plan-guid-here",
//   "service_id":     "service-guid-here",
//   "app_guid":       "app-guid-here"
// }
func BindInstance(p martini.Params, req *http.Request, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	resp := bindInstance(req, c, brokerDb, p["instance_id"], s)
	r.JSON(resp.GetStatusCode(), resp)
}

// DeleteInstance
// URL: /v2/service_instances/:instance_id
// Request:
// {
//   "service_id": "service-id-here"
//   "plan_id":    "plan-id-here"
// }
func DeleteInstance(p martini.Params, req *http.Request, r render.Render, brokerDb *gorm.DB, s *config.Settings, c *catalog.Catalog) {
	resp := deleteInstance(req, c, brokerDb, p["instance_id"], s)
	r.JSON(resp.GetStatusCode(), resp)
}
