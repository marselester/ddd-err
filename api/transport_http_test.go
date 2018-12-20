package api_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"

	"github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	"github.com/marselester/ddd-err/mock"
)

func TestUserService_ratelimit(t *testing.T) {
	s := api.NewService(nil)
	h := api.NewHTTPHandler(s, log.NewNopLogger(), 1)
	srv := httptest.NewServer(h)
	defer srv.Close()

	http.Get(srv.URL + "/v1/users/123")
	resp, err := http.Get(srv.URL + "/v1/users/456")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("FindUserByID status code: %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	want := ""
	if string(body) != want {
		t.Fatalf("FindUserByID body: %q, want %q", body, want)
	}
}

func TestUserService_FindUserByID_notfound(t *testing.T) {
	s := api.NewService(nil)
	h := api.NewHTTPHandler(s, log.NewNopLogger(), 100)
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/users/123")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("FindUserByID status code: %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	want := ""
	if string(body) != want {
		t.Fatalf("FindUserByID body: %q, want %q", body, want)
	}
}

func TestUserService_CreateUser_validation(t *testing.T) {
	tt := []struct {
		params     string
		statusCode int
		want       string
	}{
		{
			params:     `{}`,
			statusCode: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}` + "\n",
		},
		{
			params:     `{"username": ""}`,
			statusCode: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}` + "\n",
		},
		{
			params:     `{"username": " "}`,
			statusCode: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}` + "\n",
		},
		{
			params:     `{"username": ">_<"}`,
			statusCode: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_username","message":"Username is invalid."}}` + "\n",
		},
		{
			params:     `{"username": "bob123"}`,
			statusCode: http.StatusBadRequest,
			want:       `{"error":{"code":"conflict","message":"Username is already in use. Please choose a different username."}}` + "\n",
		},
	}

	s := api.NewService(&mock.UserStorage{})
	h := api.NewHTTPHandler(s, log.NewNopLogger(), 100)
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, tc := range tt {
		resp, err := http.Post(srv.URL+"/v1/users", "", strings.NewReader(tc.params))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != tc.statusCode {
			t.Fatalf("CreateUser status code: %d, want %d", resp.StatusCode, tc.statusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != tc.want {
			t.Fatalf("CreateUser body: %s, want %s", body, tc.want)
		}
	}
}

func TestUserService_CreateUser_dberror(t *testing.T) {
	s := api.NewService(&mock.UserStorage{
		UsernameInUseFn: func(ctx context.Context, username string) bool {
			return false
		},
		CreateUserFn: func(ctx context.Context, user *account.User) error {
			return &account.Error{
				Op:  "UserStorage.CreateUser",
				Err: fmt.Errorf("db connection failed"),
			}
		},
	})
	h := api.NewHTTPHandler(s, log.NewNopLogger(), 100)
	srv := httptest.NewServer(h)
	defer srv.Close()

	params := `{"username": "bob123"}`
	resp, err := http.Post(srv.URL+"/v1/users", "", strings.NewReader(params))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("CreateUser status code: %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	want := ""
	if string(body) != want {
		t.Fatalf("CreateUser body: %q, want %q", body, want)
	}
}
