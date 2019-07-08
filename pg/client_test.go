package pg

import (
	"os"
	"strconv"

	"github.com/jackc/pgx"
)

// client is a test wrapper for UserStorage implementation.
type client struct {
	// connConfig contains Postgres connection settings populated from the env.
	connConfig pgx.ConnConfig
	// conn is a connection to a test db to create schema.
	conn *pgx.Conn
	// sysConn is a connection to "postgres" db to drop/create a test db.
	sysConn *pgx.Conn

	user *UserStorage
}

// newClient returns configured test client.
// It parses Postgres connection settings for a test db.
func newClient() (*client, error) {
	config, err := parsePgEnv()
	if err != nil {
		return nil, err
	}

	c := client{
		connConfig: config,

		user: NewUserStorage(
			WithHost(config.Host),
			WithPort(config.Port),
			WithDatabase(config.Database),
			WithUser(config.User),
			WithPassword(config.Password),
		),
	}
	return &c, nil
}

// open creates the test database and opens a test client.
func (c *client) open() error {
	var err error

	config := c.connConfig
	config.Database = "postgres"
	if c.sysConn, err = pgx.Connect(config); err != nil {
		return err
	}
	if _, err = c.sysConn.Exec("DROP DATABASE IF EXISTS " + c.connConfig.Database); err != nil {
		return err
	}
	if _, err = c.sysConn.Exec("CREATE DATABASE " + c.connConfig.Database); err != nil {
		return err
	}

	if c.conn, err = pgx.Connect(c.connConfig); err != nil {
		return err
	}
	if _, err = c.conn.Exec(Schema); err != nil {
		return err
	}

	return c.user.Open()
}

// close closes client and drops the test database.
func (c *client) close() {
	c.user.Close()
	c.conn.Close()

	c.sysConn.Exec("DROP DATABASE IF EXISTS " + c.connConfig.Database)
	c.sysConn.Close()
}

/*
parsePgEnv parses the environment into Postgres connection config.
The following variables are supported:

	TEST_PGHOST, localhost by default
	TEST_PGPORT, 5432 by default
	TEST_PGDATABASE
	TEST_PGUSER
	TEST_PGPASSWORD

*/
func parsePgEnv() (pgx.ConnConfig, error) {
	config := pgx.ConnConfig{
		Host:     os.Getenv("TEST_PGHOST"),
		Database: os.Getenv("TEST_PGDATABASE"),
		User:     os.Getenv("TEST_PGUSER"),
		Password: os.Getenv("TEST_PGPASSWORD"),
	}

	if p := os.Getenv("TEST_PGPORT"); p != "" {
		if port, err := strconv.Atoi(p); err != nil {
			return config, err
		} else {
			config.Port = uint16(port)
		}
	}

	return config, nil
}

// mustOpenClient creates and opens a test client or panics.
func mustOpenClient() *client {
	c, err := newClient()
	if err != nil {
		panic(err)
	}
	if err := c.open(); err != nil {
		panic(err)
	}
	return c
}
