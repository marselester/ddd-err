package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"

	account "github.com/marselester/ddd-err"
)

// NewHTTPHandler attaches service API endpoints to HTTP routes in REST-style fashion.
func NewHTTPHandler(s account.UserService, logger log.Logger, qps int) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
	// limiter throttles requests that exceeded qps requests per second.
	// For example, when qps is 100, there might be max 100 requests per seconds to
	// all the API endpoints combined.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(
		rate.Every(time.Second), qps,
	))

	var ep endpoint.Endpoint
	{
		ep = makeCreateUserEndpoint(s)
		ep = limiter(ep)
		r.Methods("Post").Path("/v1/users").Handler(httptransport.NewServer(
			ep,
			decodeCreateUserReq,
			encodeResponse,
			options...,
		))
	}
	{
		ep = makeFindUserByIDEndpoint(s)
		ep = limiter(ep)
		r.Methods("Get").Path("/v1/users/{user_id}").Handler(httptransport.NewServer(
			ep,
			decodeFindUserByIDReq,
			encodeResponse,
			options...,
		))
	}
	return r
}

// decodeCreateUserReq converts HTTP request into service-domain request object CreateUserReq.
// Its error (e.g., json) is converted into HTTP response by encodeError.
func decodeCreateUserReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

// decodeFindUserByIDReq converts HTTP request into service-domain request object FindUserByIDReq.
func decodeFindUserByIDReq(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	req := FindUserByIDReq{ID: vars["user_id"]}
	return req, nil
}

// encodeResponse converts any service-domain response object, such as CreateUserResp,
// into HTTP response. Its error (e.g., json) is converted into HTTP response by encodeError.
// A service returns Error (business-logic error) that is shown to API client as is.
// Other service errors, e.g., DB connection error, must not be shown to API clients,
// they must not see what exactly went wrong on a server side (500 code should suffice).
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if resp, ok := response.(endpoint.Failer); ok && resp.Failed() != nil {
		err := resp.Failed()
		var accErr *account.Error
		if !errors.As(err, &accErr) {
			accErr = &account.Error{
				Code:    account.EInternal,
				Message: "An internal error has occurred.",
			}
		}

		switch accErr.Code {
		case account.ENotFound, account.EInvalidUserID:
			w.WriteHeader(http.StatusNotFound)
		case account.EInternal:
			w.WriteHeader(http.StatusInternalServerError)
			response = struct {
				Err *account.Error `json:"error"`
			}{accErr}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	return json.NewEncoder(w).Encode(response)
}

// encodeError converts errors returned by endpoint.Endpoint, its middleware (e.g., ratelimit),
// request decoder/response encoder (JSON serialization errors, e.g., EOF) into HTTP response.
// Business logic errors are not sent here.
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	errResp := struct {
		Err *account.Error `json:"error"`
	}{&account.Error{
		Code:    account.EInternal,
		Message: "An internal error has occurred.",
	}}

	if errors.Is(err, ratelimit.ErrLimited) {
		w.WriteHeader(http.StatusTooManyRequests)
		errResp.Err = &account.Error{
			Code:    account.ERateLimit,
			Message: "API rate limit exceeded.",
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(&errResp)
}
