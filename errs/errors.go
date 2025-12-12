package errs

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidParam   = errors.New("invalid parameter")
	ErrConnection     = errors.New("connection failed")
	ErrTimeout        = errors.New("operation timed out")
	ErrInternal       = errors.New("internal server error")
	ErrNotInitialized = errors.New("component not initialized")
)
