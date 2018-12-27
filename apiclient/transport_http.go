package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/sony/gobreaker"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
)

// NewHTTPClient returns UserService backed by an HTTP server living at the remote server.
func NewHTTPClient(baseURL string) (account.UserService, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := client{}
	var ep endpoint.Endpoint
	{
		ep = httptransport.NewClient(
			"POST",
			u,
			func(ctx context.Context, r *http.Request, request interface{}) error {
				r.URL.Path = "/v1/users"
				return httptransport.EncodeJSONRequest(ctx, r, request)
			},
			decodeHTTPCreateUserResp,
		).Endpoint()
		ep = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name: "CreateUser",
		}))(ep)
		c.createUserEndpoint = ep
	}
	{
		ep = httptransport.NewClient(
			"GET",
			u,
			func(ctx context.Context, r *http.Request, request interface{}) error {
				req := request.(api.FindUserByIDReq)
				r.URL.Path = "/v1/users/" + req.ID
				return httptransport.EncodeJSONRequest(ctx, r, request)
			},
			decodeHTTPFindUserByIDResp,
		).Endpoint()
		ep = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name: "FindUserByID",
		}))(ep)
		c.findUserByIDEndpoint = ep
	}
	return &c, nil
}

func decodeHTTPCreateUserResp(_ context.Context, r *http.Response) (interface{}, error) {
	resp := api.CreateUserResp{
		Err: &account.Error{},
	}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return resp, err
	}
	// Only errors returned by endpoint count against the circuit breaker's error count.
	switch account.ErrorCode(resp.Err) {
	case account.ERateLimit, account.EInternal:
		return resp, resp.Err
	}
	return resp, nil
}

func decodeHTTPFindUserByIDResp(_ context.Context, r *http.Response) (interface{}, error) {
	resp := api.FindUserByIDResp{
		Err: &account.Error{},
	}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return resp, err
	}
	// Only errors returned by endpoint count against the circuit breaker's error count.
	switch account.ErrorCode(resp.Err) {
	case account.ERateLimit, account.EInternal:
		return resp, resp.Err
	}
	return resp, nil
}
