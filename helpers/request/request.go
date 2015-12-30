package request
import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/cloudfoundry-community/aws-broker/helpers/response"
)

// CreateInstance
// URL: /v2/service_instances/:id
// Request:
// {
//   "service_id":        "service-guid-here",
//   "plan_id":           "plan-guid-here",
//   "organization_guid": "org-guid-here",
//   "space_guid":        "space-guid-here"
// }
type CreateRequest struct {
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
	OrganizationGuid string `json:"organization_guid"`
	SpaceGuid        string `json:"space_guid"`
}

func ExtractCreateRequest(req *http.Request) (CreateRequest, response.Response) {
	var cr CreateRequest
	if req.Body == nil {
		return cr, response.ErrNoRequestResponse
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return cr, response.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	json.Unmarshal(body, &cr)
	return cr, nil
}

// BindInstance
// URL: /v2/service_instances/:instance_id/service_bindings/:binding_id
// Request:
// {
//   "plan_id":        "plan-guid-here",
//   "service_id":     "service-guid-here",
//   "app_guid":       "app-guid-here"
// }
type BindRequest struct {
	PlanId			string `json:"plan_id"`
	ServiceId		string `json:"service_id"`
	AppGuid			string `json:"app_guid"`
}

func ExtractBindRequest(req *http.Request) (BindRequest, response.Response) {
	var br BindRequest
	if req.Body == nil {
		return br, response.ErrNoRequestResponse
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return br, response.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	json.Unmarshal(body, &br)
	return br, nil
}

// DeleteInstance
// URL: /v2/service_instances/:id
// Request:
// {
//   "service_id": "service-id-here"
//   "plan_id":    "plan-id-here"
// }
type DeleteRequest struct {
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
}


func ExtractDeleteRequest(req *http.Request) (DeleteRequest, response.Response) {
	var dr DeleteRequest
	if req.Body == nil {
		return dr, response.ErrNoRequestResponse
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return dr, response.NewErrorResponse(http.StatusBadRequest, err.Error())
	}
	json.Unmarshal(body, &dr)
	return dr, nil
}