// Package pg provides Postgres based UserStorage implementation.
package pg

import (
	"context"
	"database/sql"

	// pgx driver registers itself as being available to the database/sql package.
	_ "github.com/jackc/pgx/stdlib"

	account "github.com/marselester/ddd-err"
)

// UserStorage reprensets a Postgres storage to persist signed up users.
type UserStorage struct {
	config Config
	db     *sql.DB
}

// FindUserByID returns a user by ID or ENotFound error if user does not exist.
// Note, dbtx is optional.
func (s *UserStorage) FindUserByID(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error) {
	var row *sql.Row
	if dbtx == nil {
		row = s.db.QueryRowContext(ctx, "SELECT id, username FROM account WHERE id = $1", id)
	} else {
		row = dbtx.QueryRowContext(ctx, "SELECT id, username FROM account WHERE id = $1", id)
	}

	u := account.User{}
	err := row.Scan(&u.ID, &u.Username)
	if err == sql.ErrNoRows {
		return nil, &account.Error{
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
	const op = "UserStorage.CreateUser"
	_, err := s.db.ExecContext(ctx, "INSERT INTO account (id, username) VALUES ($1, $2)", u.ID, u.Username)
	if err != nil {
		return &account.Error{Op: op, Err: err}
	}
	return nil
}

// UpdateUser updates user details within a db transaction.
func (s *UserStorage) UpdateUser(ctx context.Context, dbtx *sql.Tx, u *account.User) error {
	const op = "UserStorage.UpdateUser"
	_, err := dbtx.ExecContext(ctx, "UPDATE account SET username=$2 WHERE id=$1", u.ID, u.Username)
	if err != nil {
		return &account.Error{Op: op, Err: err}
	}
	return nil
}

// Transact executes a function where transaction atomicity on the database is guaranteed.
// If the function is successfully completed, the changes are committed to the database.
// If there is an error, the changes are rolled back.
// The solution is borrowed from https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback.
func (s *UserStorage) Transact(ctx context.Context, atomic func(*sql.Tx) error) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		// Catch panics to ensure a Rollback happens right away.
		// Under normal circumstances a panic should not occur.
		// If we did not handle panics, the transaction would be rolled back eventually.
		// A non-commited transaction gets rolled back by the database when the client disconnects
		// or when the transaction gets garbage collected.
		// It's better to resolve the issue as quickly as possible.
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// err is non-nil; don't change it.
			tx.Rollback()
		} else {
			// err is nil; if Commit returns error, update err.
			err = tx.Commit()
		}
	}()

	err = atomic(tx)
	return err
}
