// Package account defines a domain of user accounts service.
package account

import (
	"context"
	"database/sql"
)

// User represents a customer in the system.
type User struct {
	ID       string
	Username string
}

// UserService represents a service for managing users.
type UserService interface {
	// FindUserByID returns a user by ID.
	FindUserByID(ctx context.Context, id string) (*User, error)
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *User) error
}

// UserStorage represents a storage for keeping user records.
type UserStorage interface {
	// FindUserByID returns a user by ID.
	FindUserByID(ctx context.Context, dbtx *sql.Tx, id string) (*User, error)
	// UsernameInUse looks up a user by username.
	UsernameInUse(ctx context.Context, username string) bool
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *User) error
	// UpdateUser updates a user.
	UpdateUser(ctx context.Context, dbtx *sql.Tx, user *User) error
}
