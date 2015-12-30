package response
import "net/http"

type Response interface {
	GetStatusCode() int
	GetResponseType() string
}

type ErrorResponse struct {
	StatusCode int `json:"-"`
	Description string `json:"description"`
}

func NewErrorResponse(statusCode int, description string) *ErrorResponse {
	return &ErrorResponse{statusCode, description}
}

func (resp *ErrorResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *ErrorResponse) GetResponseType() string {
	return "error"
}

var (
	ErrNoRequestResponse = NewErrorResponse(http.StatusBadRequest, "No Request Body")
)


type SuccessBindResponse struct {
	StatusCode int `json:"-"`
	Credentials map[string]string `json:"credentials"`
}

func NewSuccessBindResponse(credentials map[string]string) *SuccessBindResponse {
	return &SuccessBindResponse{http.StatusCreated, credentials}
}

func (resp *SuccessBindResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessBindResponse) GetResponseType() string {
	return "success_bind"
}

type SuccessCreateResponse struct {
	StatusCode int `json:"-"`
	Description string `json:"description"`
}

func NewSuccessCreateResponse() *SuccessCreateResponse {
	return &SuccessCreateResponse{http.StatusCreated, "The instance was created"}
}

func (resp *SuccessCreateResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessCreateResponse) GetResponseType() string {
	return "success_create"
}


type SuccessDeleteResponse struct {
	StatusCode int `json:"-"`
	Description string `json:"description"`
}

func NewSuccessDeleteResponse() *SuccessDeleteResponse {
	return &SuccessDeleteResponse{http.StatusCreated, "The instance was deleted"}
}

func (resp *SuccessDeleteResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessDeleteResponse) GetResponseType() string {
	return "success_create"
}

