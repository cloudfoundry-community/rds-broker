package request

import (
	"encoding/json"
	"github.com/18F/aws-broker/helpers/response"
	"io/ioutil"
	"net/http"
)

// Request:
// {
//   "service_id":        "service-guid-here",
//   "plan_id":           "plan-guid-here",
//   "organization_guid": "org-guid-here",
//   "space_guid":        "space-guid-here"
// }
type Request struct {
	ServiceId        string `json:"service_id" sql:"size(255)"`
	PlanId           string `json:"plan_id" sql:"size(255)"`
	OrganizationGuid string `json:"organization_guid" sql:"size(255)"`
	SpaceGuid        string `json:"space_guid" sql:"size(255)"`
}

func ExtractRequest(req *http.Request) (Request, response.Response) {
	var cr Request
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
