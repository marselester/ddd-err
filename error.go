package account

import (
	"errors"
	"fmt"
)

// Application error codes.
const (
	// Action cannot be performed.
	EConflict = "conflict"
	// Internal error.
	EInternal = "internal"
	// Entity does not exist.
	ENotFound = "not_found"
	// Too many API requests.
	ERateLimit = "rate_limit"
	// User ID validation failed.
	EInvalidUserID = "invalid_user_id"
	// Username validation failed.
	EInvalidUsername = "invalid_username"
)

// Error defines a standard application error.
type Error struct {
	// Code is a machine-readable error code.
	Code string `json:"code"`
	// Message is a human-readable message.
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorCode returns the code of the error, if available.
func ErrorCode(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
