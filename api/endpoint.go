package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	account "github.com/marselester/ddd-err"
)

// CreateUserReq collects the request parameters for the CreateUser method.
type CreateUserReq struct {
	Username string
}

// CreateUserResp collects the response values for the CreateUser method.
type CreateUserResp struct {
	Err error `json:"error,omitempty"`
}

// Failed implements endpoint.Failer.
func (r CreateUserResp) Failed() error { return r.Err }

// FindUserByIDReq collects the request parameters for the FindUserByID method.
type FindUserByIDReq struct {
	ID string
}

// FindUserByIDResp collects the response values for the FindUserByID method.
type FindUserByIDResp struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Err      error  `json:"error,omitempty"`
}

// Failed implements endpoint.Failer.
func (r FindUserByIDResp) Failed() error { return r.Err }

func makeCreateUserEndpoint(s account.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateUserReq)
		u := account.User{
			Username: req.Username,
		}
		err := s.CreateUser(ctx, &u)
		return CreateUserResp{Err: err}, nil
	}
}

func makeFindUserByIDEndpoint(s account.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FindUserByIDReq)
		u, err := s.FindUserByID(ctx, req.ID)
		if err != nil {
			return FindUserByIDResp{Err: err}, nil
		}
		return FindUserByIDResp{ID: u.ID, Username: u.Username}, nil
	}
}
