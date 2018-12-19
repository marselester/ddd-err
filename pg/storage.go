package pg

import (
	"context"
	"database/sql"

	"github.com/marselester/ddd-err"
)

type storage struct {
	db *sql.DB
}

// NewStorage returns a UserStorage backed by Postgres.
func NewStorage() account.UserStorage {
	return &storage{}
}

// FindUserByID returns a user by ID. Returns ENOTFOUND if user does not exist.
func (s *storage) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	u := account.User{}
	err := s.db.QueryRowContext(ctx, "SELECT id, username FROM users WHERE id = $1", id).Scan(
		&u.ID,
		&u.Username,
	)
	if err == sql.ErrNoRows {
		return nil, &account.Error{Code: account.ENOTFOUND}
	}
	return &u, err
}

// UsernameInUse returns true if username is already claimed.
func (s *storage) UsernameInUse(ctx context.Context, username string) bool {
	return true
}

// CreateUser creates a new user in the storage.
func (s *storage) CreateUser(ctx context.Context, u *account.User) error {
	const op = "UserStorage.CreateUser"
	_, err := s.db.ExecContext(ctx, "INSERT INTO account (id, username) VALUES ($1, $2)", u.ID, u.Username)
	if err != nil {
		return &account.Error{Op: op, Err: err}
	}
	return nil
}
