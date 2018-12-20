package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"

	"github.com/marselester/ddd-err/api"
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
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	want := ""
	if string(body) != want {
		t.Fatalf("body: %q, want %q", body, want)
	}
	if resp.StatusCode != 429 {
		t.Fatalf("status code: %d, want %d", resp.StatusCode, 429)
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
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	want := ""
	if string(body) != want {
		t.Fatalf("body: %q, want %q", body, want)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("status code: %d, want %d", resp.StatusCode, 404)
	}
}

func TestUserService_CreateUser_error(t *testing.T) {
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
	}

	s := api.NewService(nil)
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
