package account

import (
	"bytes"
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

	// Logical operation and nested error.
	Op  string `json:"-"`
	Err error  `json:"-"`
}

// Error returns the string representation of the error message.
// This implementation assumes that Err cannot coexist with Code or Message on any given error.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Print the current operation in our stack, if any.
	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message.
	// Otherwise print the error code & message.
	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "%s: ", e.Code)
		}
		buf.WriteString(e.Message)
	}
	return buf.String()
}

// ErrorCode returns the code of the root error, if available. Otherwise returns EInternal.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return ErrorCode(e.Err)
	}
	return EInternal
}

// ErrorMessage returns the human-readable message of the error, if available.
// Otherwise returns a generic error message.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok && e.Message != "" {
		return e.Message
	} else if ok && e.Err != nil {
		return ErrorMessage(e.Err)
	}
	return "An internal error has occurred. Please contact technical support."
}
