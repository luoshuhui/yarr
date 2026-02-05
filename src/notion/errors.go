package notion

import "errors"

var (
	ErrMissingAPIKey     = errors.New("notion: API key is required")
	ErrMissingDatabaseID = errors.New("notion: database ID is required")
	ErrAPICallFailed     = errors.New("notion: API call failed")
	ErrConversionFailed  = errors.New("notion: HTML conversion failed")
)
