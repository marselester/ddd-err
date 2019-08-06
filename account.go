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

// Group represents a group of customers.
type Group struct {
	ID   string
	Name string
}

// UserService represents a service for managing users.
type UserService interface {
	// FindUserByID returns a user by ID.
	FindUserByID(ctx context.Context, id string) (*User, error)
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *User) error
}

// Storage allows repositories to execute SQL transactions.
// For example, a service might need to call CreateUser and CreateGroup within the same Postgres transaction.
type Storage interface {
	Transact(ctx context.Context, atomic func(*sql.Tx) error) error
}

// UserRepository represents a storage for keeping user records.
type UserRepository interface {
	Storage
	// FindUserByID returns a user by ID.
	FindUserByID(ctx context.Context, dbtx *sql.Tx, id string) (*User, error)
	// UsernameInUse looks up a user by username.
	UsernameInUse(ctx context.Context, username string) bool
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *User) error
	// UpdateUser updates a user.
	UpdateUser(ctx context.Context, dbtx *sql.Tx, user *User) error
}

// GroupRepository represents a storage for keeping customer group records.
type GroupRepository interface {
	Storage
	// CreateGroup creates a new group.
	CreateGroup(ctx context.Context, dbtx *sql.Tx, group *Group) error
}
