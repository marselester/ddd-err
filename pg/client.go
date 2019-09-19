package pg

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx"
	// pgx driver registers itself as being available to the database/sql package.
	pgxstdlib "github.com/jackc/pgx/stdlib"
)

// Client represents a client to the underlying PostgreSQL data store.
type Client struct {
	User *UserStorage

	config Config
	db     *sql.DB
}

// NewClient returns a new Postgres client.
// By default there are 5 max simultaneous Postgres connections.
func NewClient(options ...ConfigOption) *Client {
	c := Client{
		config: Config{
			host:           "localhost",
			port:           5432,
			maxConnections: 5,
		},
	}
	c.User = &UserStorage{client: &c}

	for _, opt := range options {
		opt(&c.config)
	}
	return &c
}

// Open connects to a PostgreSQL DB.
func (c *Client) Open() error {
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     c.config.host,
			Port:     c.config.port,
			Database: c.config.database,
			User:     c.config.user,
			Password: c.config.password,
		},
		MaxConnections: c.config.maxConnections,
	}
	pool, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		return err
	}

	c.db = pgxstdlib.OpenDBFromPool(pool)
	return nil
}

// Close closes PostgreSQL connection.
func (c *Client) Close() error {
	return c.db.Close()
}

// Transact executes a function where transaction atomicity on the database is guaranteed.
// If the function is successfully completed, the changes are committed to the database.
// If there is an error, the changes are rolled back.
// The solution is borrowed from https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback.
func (c *Client) Transact(ctx context.Context, atomic func(*sql.Tx) error) (err error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		// Catch panics to ensure a Rollback happens right away.
		// Under normal circumstances a panic should not occur.
		// If we did not handle panics, the transaction would be rolled back eventually.
		// A non-committed transaction gets rolled back by the database when the client disconnects
		// or when the transaction gets garbage collected.
		// It's better to resolve the issue as quickly as possible.
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
		// err is not nil; don't change it.
		if err != nil {
			tx.Rollback()
			return
		}
		// err is nil; if Commit returns error, update err.
		err = tx.Commit()
	}()

	err = atomic(tx)
	return err
}
