package api

import (
	"context"

	"github.com/marselester/ddd-err"
)

type service struct {
	storage account.UserStorage
}

// NewService returns a service that manages user accounts.
func NewService(storage account.UserStorage) account.UserService {
	return &service{
		storage: storage,
	}
}

func (s *service) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	if id == "" {
		return nil, &account.Error{
			Code:    account.EINVALID,
			Message: "ID is required.",
		}
	}

	u, err := s.storage.FindUserByID(ctx, id)
	if account.ErrorCode(err) == account.ENOTFOUND {
		// Log event...
	}
	return u, err
}

// CreateUser creates a new user in the system.
// Returns EINVALID if the username is blank or already exists.
// Returns ECONFLICT if the username is already in use.
func (s *service) CreateUser(ctx context.Context, u *account.User) error {
	// Validate username is non-blank.
	if u.Username == "" {
		return &account.Error{
			Code:    account.EINVALID,
			Message: "Username is required.",
		}
	}

	// Verify user does not already exist.
	if s.storage.UsernameInUse(ctx, u.Username) {
		return &account.Error{
			Code:    account.ECONFLICT,
			Message: "Username is already in use. Please choose a different username.",
		}
	}

	return s.storage.CreateUser(ctx, u)
}
