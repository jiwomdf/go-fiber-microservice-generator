package helper

// Response is the standardized JSON response structure.
type Response struct {
	Status Status      `json:"status"`
	Meta   interface{} `json:"meta,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// Status represents the status part of the response.
type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewResponse creates a new standardized response object.
func NewResponse(code string, message string, meta interface{}, data interface{}) Response {
	return Response{
		Status: Status{Code: code, Message: message},
		Meta:   meta,
		Data:   data,
	}
}
