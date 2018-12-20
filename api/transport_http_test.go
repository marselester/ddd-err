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

func TestUserService_CreateUser(t *testing.T) {
	s := api.NewService(nil)
	h := api.NewHTTPHandler(s, log.NewNopLogger(), 100)
	srv := httptest.NewServer(h)
	defer srv.Close()

	params := strings.NewReader(`{}`)
	resp, err := http.Post(srv.URL+"/v1/users", "", params)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	want := `{"error":{"code":"invalid_username","message":"Username is required."}}`
	if strings.TrimSpace(string(body)) != want {
		t.Fatalf("body: %s, want %s", body, want)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("status code: %d, want %d", resp.StatusCode, 400)
	}
}
