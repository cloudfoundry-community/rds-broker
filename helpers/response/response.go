package response

import "net/http"

// Response represents the common data for all types of responses to have and implement.
type Response interface {
	GetStatusCode() int
	GetResponseType() Type
}

type errorResponse struct {
	StatusCode  int    `json:"-"`
	Description string `json:"description"` // Necessary for all error responses.
}

// Type indicates the type of response. Nice for debug situations.
type Type string

// These contain the list of response types. Useful for debug situations.
var (
	// SuccessCreateResponseType represents a response for a successful instance creation.
	SuccessCreateResponseType Type = "success_create"
	// SuccessBindResponseType represents a response for a successful instance binding.
	SuccessBindResponseType Type = "success_bind"
	// SuccessDeleteResponseType represents a response for a successful instance deletion.
	SuccessDeleteResponseType Type = "success_delete"
	// ErrorResponseType represents a response for an error.
	ErrorResponseType Type = "error"
)

// NewErrorResponse is the constructor for an errorResponse.
func NewErrorResponse(statusCode int, description string) Response {
	return &errorResponse{statusCode, description}
}

func (resp *errorResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *errorResponse) GetResponseType() Type {
	return ErrorResponseType
}

var (
	// ErrNoRequestBodyResponse is a response indicating there was no request body.
	ErrNoRequestBodyResponse = NewErrorResponse(http.StatusBadRequest, "No Request Body")
)

type successBindResponse struct {
	StatusCode  int               `json:"-"`
	Credentials map[string]string `json:"credentials"` // Needed for sending credentials for service.
}

// NewSuccessBindResponse is the constructor for a successBindResponse.
func NewSuccessBindResponse(credentials map[string]string) Response {
	return &successBindResponse{http.StatusCreated, credentials}
}

func (resp *successBindResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *successBindResponse) GetResponseType() Type {
	return SuccessBindResponseType
}

type genericSuccessResponse struct {
	StatusCode  int    `json:"-"`
	Description string `json:"description"`
	StatusType Type `json:"-"`

}

func (resp *genericSuccessResponse) GetStatusCode() int {
	return resp.StatusCode
}

func (resp *genericSuccessResponse) GetResponseType() Type {
	return SuccessCreateResponseType
}

var (
	// SuccessCreateResponse represents the response that all successful instance creations should return.
	SuccessCreateResponse = &genericSuccessResponse{http.StatusCreated, "The instance was created", SuccessCreateResponseType}
	// SuccessDeleteResponse represents the response that all successful instance deletions should return.
	SuccessDeleteResponse = &genericSuccessResponse{http.StatusOK, "The instance was deleted", SuccessDeleteResponseType}
)

