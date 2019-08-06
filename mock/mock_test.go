package mock

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	account "github.com/marselester/ddd-err"
)

// TestStorageTransact is a dummy example of mocking a service, repository and db transaction.
func TestStorageTransact(t *testing.T) {
	userRepo := UserStorage{FindUserByIDFn: func(_ context.Context, _ *sql.Tx, _ string) (*account.User, error) {
		return nil, &account.Error{
			Code:    account.ENotFound,
			Message: "User not found.",
		}
	}}
	groupRepo := GroupStorage{}

	s := UserService{CreateUserFn: func(ctx context.Context, user *account.User) error {
		err := userRepo.Transact(ctx, func(tx *sql.Tx) error {
			_, err := userRepo.FindUserByID(ctx, tx, user.ID)
			if account.ErrorCode(err) != account.ENotFound {
				return &account.Error{
					Code:    "shoe_fell_off",
					Message: "Username is already in use. Please choose a different username.",
				}
			}

			if err = userRepo.CreateUser(ctx, user); err != nil {
				return err
			}

			group := account.Group{
				ID:   user.ID,
				Name: strings.ToLower(user.Username),
			}
			return groupRepo.CreateGroup(ctx, tx, &group)
		})

		return err
	}}

	alice := account.User{
		ID:       "123",
		Username: "Alice",
	}
	err := s.CreateUser(context.Background(), &alice)
	if err != nil {
		t.Errorf("CreateUser() failed: %v", err)
	}

	userRepo.FindUserByIDFn = func(_ context.Context, _ *sql.Tx, _ string) (*account.User, error) {
		return nil, fmt.Errorf("shoe fell off")
	}
	err = s.CreateUser(context.Background(), &alice)
	code := account.ErrorCode(err)
	if code != "shoe_fell_off" {
		t.Errorf("CreateUser() got %q error code, want shoe_fell_off", code)
	}
}
