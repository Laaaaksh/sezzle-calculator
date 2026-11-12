package apperror

// Public is the client-visible part of an error: the code and public message,
// and nothing about the cause.
type Public struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// ToPublic extracts the response body for err.
func ToPublic(err IError) Public {
	return Public{Code: err.Code(), Message: err.PublicMessage()}
}
