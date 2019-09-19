package apiclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/apiclient"
)

func TestUserService_CreateUser_errors(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		statusCode int
		want       string
	}{
		{
			name:       "invalid username",
			body:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}`,
			statusCode: http.StatusBadRequest,
			want:       "invalid_username: Username is invalid.",
		},
		{
			name:       "username conflict",
			body:       `{"error":{"code":"conflict","message":"Username is already in use. Please choose a different username."}}`,
			statusCode: http.StatusBadRequest,
			want:       "conflict: Username is already in use. Please choose a different username.",
		},
		{
			name:       "too many requests",
			body:       `{"error":{"code":"rate_limit","message":"API rate limit exceeded."}}`,
			statusCode: http.StatusTooManyRequests,
			want:       "rate_limit: API rate limit exceeded.",
		},
		{
			name:       "server error",
			body:       `{"error":{"code":"internal","message":"An internal error has occurred."}}`,
			statusCode: http.StatusInternalServerError,
			want:       "internal: An internal error has occurred.",
		},
		{
			name:       "json error",
			body:       "{",
			statusCode: http.StatusOK,
			want:       "unexpected EOF",
		},
		{
			name:       "empty response",
			body:       "",
			statusCode: http.StatusInternalServerError,
			want:       "EOF",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tc.body, tc.statusCode)
			}))
			defer srv.Close()

			c, err := apiclient.NewHTTPClient(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.CreateUser(context.Background(), &account.User{})
			if err.Error() != tc.want {
				t.Errorf("CreateUser error %q, want %q", err, tc.want)
			}
		})
	}
}

func TestUserService_FindUserByID_errors(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		statusCode int
		want       string
	}{
		{
			name:       "invalid user id",
			body:       `{"error":{"code":"invalid_user_id","message":"Invalid user ID."}}`,
			statusCode: http.StatusNotFound,
			want:       "invalid_user_id: Invalid user ID.",
		},
		{
			name:       "user not found",
			body:       `{"error":{"code":"not_found","message":"User not found."}}`,
			statusCode: http.StatusNotFound,
			want:       "not_found: User not found.",
		},
		{
			name:       "server error",
			body:       `{"error":{"code":"internal","message":"An internal error has occurred."}}`,
			statusCode: http.StatusInternalServerError,
			want:       "internal: An internal error has occurred.",
		},
		{
			name:       "empty response",
			body:       "",
			statusCode: http.StatusInternalServerError,
			want:       "EOF",
		},
		{
			name:       "json error",
			body:       "{",
			statusCode: http.StatusOK,
			want:       "unexpected EOF",
		},
		{
			name:       "too many requests",
			body:       `{"error":{"code":"rate_limit","message":"API rate limit exceeded."}}`,
			statusCode: http.StatusTooManyRequests,
			want:       "rate_limit: API rate limit exceeded.",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tc.body, tc.statusCode)
			}))
			defer srv.Close()

			c, err := apiclient.NewHTTPClient(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			_, err = c.FindUserByID(context.Background(), "123")
			if err.Error() != tc.want {
				t.Errorf("FindUserByID error %q, want %q", err, tc.want)
			}
		})
	}
}

func TestUserService_circuitbreaker(t *testing.T) {
	tt := []struct {
		name       string
		body       string
		statusCode int
		want       string
	}{
		{
			name:       "validation error",
			body:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}`,
			statusCode: http.StatusBadRequest,
			want:       "invalid_username: Username is invalid.",
		},
		{
			name:       "too many requests",
			body:       `{"error":{"code":"rate_limit","message":"API rate limit exceeded."}}`,
			statusCode: http.StatusTooManyRequests,
			want:       "circuit breaker is open",
		},
		{
			name:       "server error",
			body:       `{"error":{"code":"internal","message":"An internal error has occurred."}}`,
			statusCode: http.StatusInternalServerError,
			want:       "circuit breaker is open",
		},
		{
			name:       "json error",
			body:       "",
			statusCode: http.StatusInternalServerError,
			want:       "circuit breaker is open",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tc.body, tc.statusCode)
			}))
			defer srv.Close()

			c, err := apiclient.NewHTTPClient(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 7; i++ {
				err = c.CreateUser(context.Background(), &account.User{})
			}
			if err.Error() != tc.want {
				t.Errorf("CreateUser error %q, want %q", err, tc.want)
			}
		})
	}
}
