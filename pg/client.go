package pg

import (
	"github.com/jackc/pgx"
	// pgx driver registers itself as being available to the database/sql package.
	pgxstdlib "github.com/jackc/pgx/stdlib"
)

// NewUserStorage returns a UserStorage backed by Postgres.
// By default there are 5 max simultaneous Postgres connections.
func NewUserStorage(options ...ConfigOption) *UserStorage {
	s := UserStorage{
		config: Config{
			host:           "localhost",
			port:           5432,
			maxConnections: 5,
		},
	}

	for _, opt := range options {
		opt(&s.config)
	}
	return &s
}

// Open connects to a PostgreSQL DB.
func (s *UserStorage) Open() error {
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     s.config.host,
			Port:     s.config.port,
			Database: s.config.database,
			User:     s.config.user,
			Password: s.config.password,
		},
		MaxConnections: s.config.maxConnections,
	}
	pool, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		return err
	}

	s.db = pgxstdlib.OpenDBFromPool(pool)
	return nil
}

// Close closes PostgreSQL connection.
func (s *UserStorage) Close() error {
	return s.db.Close()
}
