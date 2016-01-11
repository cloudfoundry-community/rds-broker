package base

import (
	"github.com/18F/aws-broker/catalog"
	"github.com/18F/aws-broker/helpers/request"
	"github.com/18F/aws-broker/helpers/response"
)

type Broker interface {
	CreateInstance(*catalog.Catalog, string, request.Request) response.Response
	BindInstance(*catalog.Catalog, string, Instance) response.Response
	DeleteInstance(*catalog.Catalog, string, Instance) response.Response
}
