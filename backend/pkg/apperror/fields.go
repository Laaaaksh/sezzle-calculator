package apperror

// Field keys for the "fields" object of an error body. They are wire names
// that clients read, so they are kept apart from the log keys in
// internal/constants, which may be renamed without breaking anyone.
const (
	FieldMethod = "method"
	FieldPath   = "path"
	FieldLimit  = "limit"
)
