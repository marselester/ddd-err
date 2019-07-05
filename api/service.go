// Package api provides REST-style API server for managing users.
package api

import (
	"context"
	"regexp"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"

	account "github.com/marselester/ddd-err"
)

type service struct {
	logger log.Logger
	db     account.UserStorage
}

// FindUserByID returns a user by its ID.
// It returns EInvalidUserID if the ID is invalid UUID.
func (s *service) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, &account.Error{
			Code:    account.EInvalidUserID,
			Message: "Invalid user ID.",
		}
	}

	return s.db.FindUserByID(ctx, nil, userID.String())
}

var validUsername = regexp.MustCompile(`^[A-z0-9]+$`)

// CreateUser creates a new user in the system.
// It returns EInvalidUsername if the username is blank or
// EConflict if the username is already in use.
func (s *service) CreateUser(ctx context.Context, u *account.User) error {
	if !validUsername.MatchString(u.Username) {
		return &account.Error{
			Code:    account.EInvalidUsername,
			Message: "Username is invalid.",
		}
	}

	if s.db.UsernameInUse(ctx, u.Username) {
		return &account.Error{
			Code:    account.EConflict,
			Message: "Username is already in use. Please choose a different username.",
		}
	}

	return s.db.CreateUser(ctx, u)
}
