// Package mock provides mocks for the account domain to facilitate testing.
package mock

import (
	"context"

	"github.com/marselester/ddd-err"
)

// UserService is a mock that implements account.UserService.
type UserService struct {
	FindUserByIDFn     func(ctx context.Context, id string) (*account.User, error)
	FindUserByIDCalled bool
	CreateUserFn       func(ctx context.Context, user *account.User) error
	CreateUserCalled   bool
}

// FindUserByID calls FindUserByIDFn and sets FindUserByIDCalled = true for tests
// to inspect the mock.
func (s *UserService) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	s.FindUserByIDCalled = true
	if s.FindUserByIDFn == nil {
		return &account.User{}, nil
	}
	return s.FindUserByIDFn(ctx, id)
}

// CreateUser calls CreateUserFn and sets CreateUserCalled = true for tests to inspect the mock.
func (s *UserService) CreateUser(ctx context.Context, user *account.User) error {
	s.CreateUserCalled = true
	if s.CreateUserFn == nil {
		return nil
	}
	return s.CreateUserFn(ctx, user)
}
