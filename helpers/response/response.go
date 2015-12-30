package response

type Response struct {
	StatusCode int `json:"-"`
	Description string `json:"description"`
}

func New(statusCode int, description string) *Response {
	return &Response{statusCode, description}
}
