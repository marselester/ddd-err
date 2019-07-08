package pg

import (
	"context"
	"database/sql"
	"testing"

	account "github.com/marselester/ddd-err"
)

// Ensure UserStorage implements account.UserStorage.
var _ account.UserStorage = &UserStorage{}

func TestTransact(t *testing.T) {
	c := mustOpenClient()
	defer c.close()

	ctx := context.Background()
	alice := account.User{
		ID:       "123",
		Username: "Alice",
	}
	c.user.CreateUser(ctx, &alice)

	err := c.user.Transact(ctx, func(tx *sql.Tx) error {
		acc, err := c.user.FindUserByID(ctx, tx, alice.ID)
		if err != nil {
			return err
		}

		acc.Username = "Bob"
		return c.user.UpdateUser(ctx, tx, acc)
	})
	if err != nil {
		t.Errorf("Transact() failed: %v", err)
	}

	bob, err := c.user.FindUserByID(ctx, nil, alice.ID)
	if err != nil {
		t.Errorf("Transact() user not found by ID %q: %v", alice.ID, err)
	}
	if bob.Username != "Bob" {
		t.Errorf("Transact() got username %q, want Bob", bob.Username)
	}
}
