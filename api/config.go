package api

import (
	"github.com/go-kit/kit/log"

	account "github.com/marselester/ddd-err"
)

// NewService configures new UserService that manages user accounts.
// You must provide a repository where users are stored.
func NewService(db account.UserRepository, options ...ConfigOption) account.UserService {
	s := service{
		logger: log.NewNopLogger(),
		db:     db,
	}
	for _, opt := range options {
		opt(&s)
	}
	return &s
}

// ConfigOption configures the UserService.
type ConfigOption func(*service)

// WithLogger configures a logger to debug the service.
func WithLogger(l log.Logger) ConfigOption {
	return func(r *service) {
		r.logger = l
	}
}
