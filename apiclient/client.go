// Package apiclient provides API client for managing users.
package apiclient

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
)

// client represents an API client for UserService backed by remote server.
type client struct {
	findUserByIDEndpoint endpoint.Endpoint
	createUserEndpoint   endpoint.Endpoint
}

// FindUserByID requests user info by ID from API server.
func (c *client) FindUserByID(ctx context.Context, id string) (*account.User, error) {
	req := api.FindUserByIDReq{ID: id}
	response, err := c.findUserByIDEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	resp := response.(api.FindUserByIDResp)
	if resp.Err != nil {
		return nil, resp.Err
	}
	u := account.User{
		ID:       resp.ID,
		Username: resp.Username,
	}
	return &u, nil
}

// CreateUser creates user at API server.
func (c *client) CreateUser(ctx context.Context, user *account.User) error {
	req := api.CreateUserReq{Username: user.Username}
	response, err := c.createUserEndpoint(ctx, req)
	if err != nil {
		return err
	}
	resp := response.(api.CreateUserResp)
	return resp.Err
}
