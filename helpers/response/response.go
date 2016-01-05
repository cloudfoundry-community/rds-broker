package response

import "net/http"

type Response interface {
	GetStatusCode() int
	GetResponseType() ResponseType
}

type ErrorResponse struct {
	StatusCode  int    `json:"-"`
	Description string `json:"description"`
}

type ResponseType string

var (
	SuccessCreateResponseType ResponseType = "success_create"
	SuccessBindResponseType   ResponseType = "success_bind"
	SuccessDeleteResponseType ResponseType = "success_delete"
	ErrorResponseType         ResponseType = "error"
)

func NewErrorResponse(statusCode int, description string) *ErrorResponse {
	return &ErrorResponse{statusCode, description}
}

func (resp *ErrorResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *ErrorResponse) GetResponseType() ResponseType {
	return ErrorResponseType
}

var (
	ErrNoRequestResponse = NewErrorResponse(http.StatusBadRequest, "No Request Body")
)

type SuccessBindResponse struct {
	StatusCode  int               `json:"-"`
	Credentials map[string]string `json:"credentials"`
}

func NewSuccessBindResponse(credentials map[string]string) *SuccessBindResponse {
	return &SuccessBindResponse{http.StatusCreated, credentials}
}

func (resp *SuccessBindResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessBindResponse) GetResponseType() ResponseType {
	return SuccessBindResponseType
}

type SuccessCreateResponse struct {
	StatusCode  int    `json:"-"`
	Description string `json:"description"`
}

func NewSuccessCreateResponse() *SuccessCreateResponse {
	return &SuccessCreateResponse{http.StatusCreated, "The instance was created"}
}

func (resp *SuccessCreateResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessCreateResponse) GetResponseType() ResponseType {
	return SuccessCreateResponseType
}

type SuccessDeleteResponse struct {
	StatusCode  int    `json:"-"`
	Description string `json:"description"`
}

func NewSuccessDeleteResponse() *SuccessDeleteResponse {
	return &SuccessDeleteResponse{http.StatusOK, "The instance was deleted"}
}

func (resp *SuccessDeleteResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *SuccessDeleteResponse) GetResponseType() ResponseType {
	return SuccessDeleteResponseType
}
