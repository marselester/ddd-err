package api

import (
	"context"
	"time"

	"github.com/go-kit/log"

	account "github.com/marselester/ddd-err"
)

// NewLoggingMiddleware makes a logging middleware for UserService that
// logs user creation attempts, and which errors occurred (invalid username format,
// storage connection errors).
func NewLoggingMiddleware(l log.Logger, s account.UserService) account.UserService {
	return &loggingMiddleware{
		logger: l,
		next:   s,
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   account.UserService
}

func (mw *loggingMiddleware) FindUserByID(ctx context.Context, id string) (v *account.User, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "FindUserByID",
			"user_id", id,
			"output", v,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	v, err = mw.next.FindUserByID(ctx, id)
	return
}

func (mw *loggingMiddleware) CreateUser(ctx context.Context, user *account.User) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "CreateUser",
			"user", user,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.next.CreateUser(ctx, user)
	return
}
