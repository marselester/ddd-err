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
	// Inner is a wrapped error that is never shown to API consumers.
	Inner error `json:"-"`
}

func (e Error) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Inner)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e Error) Unwrap() error {
	return e.Inner
}

// ErrorCode returns the code of the error, if available.
func ErrorCode(err error) string {
	var e Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
