package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/marselester/ddd-err"
)

type createUserReq struct {
	Username string
}
type createUserResp struct {
	Err error `json:"error,omitempty"`
}

func (r createUserResp) Failed() error { return r.Err }

type findUserByIDReq struct {
	ID string
}
type findUserByIDResp struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Err      error  `json:"error,omitempty"`
}

func (r findUserByIDResp) Failed() error { return r.Err }

func makeCreateUserEndpoint(s account.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createUserReq)
		u := account.User{
			Username: req.Username,
		}
		err := s.CreateUser(ctx, &u)
		return createUserResp{Err: err}, nil
	}
}

func makeFindUserByIDEndpoint(s account.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findUserByIDReq)
		u, err := s.FindUserByID(ctx, req.ID)
		if err != nil {
			return findUserByIDResp{Err: err}, nil
		}
		return findUserByIDResp{ID: u.ID, Username: u.Username}, nil
	}
}
