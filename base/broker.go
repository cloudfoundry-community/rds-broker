package base

import (
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/cloudfoundry-community/aws-broker/catalog"
)

type Broker interface {
	CreateInstance(catalog.AWSPlan, string) *response.Response
	BindInstance(catalog.AWSPlan, string) *response.Response
	UnbindInstance() *response.Response
	DeleteInstance() *response.Response
}