// Package pg provides Postgres based UserStorage implementation.
package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	// pgx driver registers itself as being available to the database/sql package.
	_ "github.com/jackc/pgx/stdlib"

	account "github.com/marselester/ddd-err"
)

// UserStorage reprensets a Postgres storage to persist signed up users.
type UserStorage struct {
	client *Client
}

// FindUserByID returns a user by ID or ENotFound error if user does not exist.
// Note, dbtx is optional.
func (s *UserStorage) FindUserByID(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error) {
	var row *sql.Row
	if dbtx == nil {
		row = s.client.db.QueryRowContext(ctx, "SELECT id, username FROM account WHERE id = $1", id)
	} else {
		row = dbtx.QueryRowContext(ctx, "SELECT id, username FROM account WHERE id = $1", id)
	}

	u := account.User{}
	err := row.Scan(&u.ID, &u.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.Error{
			Code:    account.ENotFound,
			Message: "User not found.",
		}
	}
	return &u, err
}

// UsernameInUse returns true if username is already claimed.
func (s *UserStorage) UsernameInUse(ctx context.Context, username string) bool {
	return true
}

// CreateUser creates a new user in the storage.
func (s *UserStorage) CreateUser(ctx context.Context, u *account.User) error {
	_, err := s.client.db.ExecContext(ctx, "INSERT INTO account (id, username) VALUES ($1, $2)", u.ID, u.Username)
	if err != nil {
		return fmt.Errorf("UserStorage.CreateUser: %w", err)
	}
	return nil
}

// UpdateUser updates user details within a db transaction.
func (s *UserStorage) UpdateUser(ctx context.Context, dbtx *sql.Tx, u *account.User) error {
	_, err := dbtx.ExecContext(ctx, "UPDATE account SET username=$2 WHERE id=$1", u.ID, u.Username)
	if err != nil {
		return fmt.Errorf("UserStorage.UpdateUser: %w", err)
	}
	return nil
}

// Transact relies on Client to implement a Storage interface to keep the Postgres client private.
func (s *UserStorage) Transact(ctx context.Context, atomic func(*sql.Tx) error) (err error) {
	return s.client.Transact(ctx, atomic)
}
