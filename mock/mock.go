// Package mock provides mocks for the account domain to facilitate testing.
package mock

import (
	"context"

	account "github.com/marselester/ddd-err"
)

// UserService is a mock that implements account.UserService.
type UserService struct {
	FindUserByIDFn func(ctx context.Context, id string) (*account.User, error)
	CreateUserFn   func(ctx context.Context, user *account.User) error
}

// FindUserByID calls FindUserByIDFn for tests to inspect the mock.
func (s *UserService) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	if s.FindUserByIDFn == nil {
		return &account.User{}, nil
	}
	return s.FindUserByIDFn(ctx, id)
}

// CreateUser calls CreateUserFn for tests to inspect the mock.
func (s *UserService) CreateUser(ctx context.Context, user *account.User) error {
	if s.CreateUserFn == nil {
		return nil
	}
	return s.CreateUserFn(ctx, user)
}

// UserStorage is a mock that implements account.UserStorage.
type UserStorage struct {
	FindUserByIDFn  func(ctx context.Context, id string) (*account.User, error)
	UsernameInUseFn func(ctx context.Context, username string) bool
	CreateUserFn    func(ctx context.Context, user *account.User) error
}

// FindUserByID calls FindUserByIDFn for tests to inspect the mock.
func (s *UserStorage) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	if s.FindUserByIDFn == nil {
		return &account.User{}, nil
	}
	return s.FindUserByIDFn(ctx, id)
}

// UsernameInUse calls UsernameInUseFn for tests to inspect the mock.
func (s *UserStorage) UsernameInUse(ctx context.Context, username string) bool {
	if s.UsernameInUseFn == nil {
		return true
	}
	return s.UsernameInUseFn(ctx, username)
}

// CreateUser calls CreateUserFn for tests to inspect the mock.
func (s *UserStorage) CreateUser(ctx context.Context, user *account.User) error {
	if s.CreateUserFn == nil {
		return nil
	}
	return s.CreateUserFn(ctx, user)
}
