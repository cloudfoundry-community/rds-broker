package base

import (
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
	"github.com/cloudfoundry-community/aws-broker/catalog"
"github.com/cloudfoundry-community/aws-broker/helpers/request"
)

type Broker interface {
	CreateInstance(catalog.AWSPlan, string, request.CreateRequest) response.Response
	BindInstance(catalog.AWSPlan, string) response.Response
	DeleteInstance(catalog.AWSPlan, string) response.Response
}