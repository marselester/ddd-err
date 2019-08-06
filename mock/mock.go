// Package mock provides mocks for the account domain to facilitate testing.
package mock

import (
	"context"
	"database/sql"

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

// Storage is a mock that implements account.Storage.
type Storage struct {
	TransactFn func(ctx context.Context, atomic func(*sql.Tx) error) error
}

// Transact calls TransactFn for tests to inspect the mock.
func (s *Storage) Transact(ctx context.Context, atomic func(*sql.Tx) error) (err error) {
	if s.TransactFn == nil {
		return atomic(nil)
	}
	return s.TransactFn(ctx, atomic)
}

// UserStorage is a mock that implements account.UserRepository.
type UserStorage struct {
	Storage
	FindUserByIDFn  func(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error)
	UsernameInUseFn func(ctx context.Context, username string) bool
	CreateUserFn    func(ctx context.Context, user *account.User) error
	UpdateUserFn    func(ctx context.Context, dbtx *sql.Tx, user *account.User) error
}

// FindUserByID calls FindUserByIDFn for tests to inspect the mock.
func (s *UserStorage) FindUserByID(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error) {
	if s.FindUserByIDFn == nil {
		return &account.User{}, nil
	}
	return s.FindUserByIDFn(ctx, dbtx, id)
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

// UpdateUser calls UpdateUserFn for tests to inspect the mock.
func (s *UserStorage) UpdateUser(ctx context.Context, dbtx *sql.Tx, user *account.User) error {
	if s.UpdateUserFn == nil {
		return nil
	}
	return s.UpdateUserFn(ctx, dbtx, user)
}

// GroupStorage is a mock that implements account.GroupRepository.
type GroupStorage struct {
	Storage
	CreateGroupFn func(ctx context.Context, dbtx *sql.Tx, group *account.Group) error
}

// CreateGroup calls CreateGroupFn for tests to inspect the mock.
func (s *GroupStorage) CreateGroup(ctx context.Context, dbtx *sql.Tx, group *account.Group) error {
	if s.CreateGroupFn == nil {
		return nil
	}
	return s.CreateGroupFn(ctx, dbtx, group)
}
