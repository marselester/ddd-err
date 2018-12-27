# DDD Error Handling

[![Documentation](https://godoc.org/github.com/marselester/ddd-err?status.svg)](https://godoc.org/github.com/marselester/ddd-err)
[![Go Report Card](https://goreportcard.com/badge/github.com/marselester/ddd-err)](https://goreportcard.com/report/github.com/marselester/ddd-err)

This is an error handling example based on Ben Johnson's
[Failure is your Domain](https://middlemost.com/failure-is-your-domain/).
According to the article error consumers have different expectations:

- end user requires human-readable message. API undefined errors must not be
  shown, e.g., Postgres error can reveal db schema.
- app itself can recover from error states using error codes.
- operator should be able to debug and see all errors including stack trace.

Let's start the API server and see that in action.

```sh
$ go run ./cmd/server/
```

Domain errors (API errors) should have `Code` and `Message`. For instance, "duplicate username" error

```go
&account.Error{
	Code:    account.EConflict,
	Message: "Username is already in use. Please choose a different username.",
}
```

is shown to the API consumer

```sh
$ curl -i -X POST -d '{"username":"bob"}' http://localhost:8000/v1/users
HTTP/1.1 400 Bad Request

{"error":{"code":"conflict","message":"Username is already in use. Please choose a different username."}}
```

and also logged for operators

```json
{
  "caller": "middleware.go:44",
  "err": "conflict: Username is already in use. Please choose a different username.",
  "method": "CreateUser",
  "took": "2.257µs",
  "ts": "2018-12-20T13:49:10.379131Z",
  "user": {
    "ID": "",
    "Username": "bob"
  }
}
```

Errors are considered internal and not shown to API consumers:

- third-party errors, e.g., `fmt.Errorf("oh no")`
- `account.Error` with blank `Code` field

For example, db connection error

```go
&account.Error{
	Op: "UserStorage.CreateUser",
	Err: &account.Error{
		Op:  "insertUser",
		Err: fmt.Errorf("db connection failed"),
	},
}
```

is suppressed on API level

```sh
$ curl -i -X POST -d '{"username":"alice"}' http://localhost:8000/v1/users
HTTP/1.1 500 Internal Server Error

{"error":{"code":"internal","message":"An internal error has occurred."}}
```

but logged for operators

```json
{
  "caller": "middleware.go:44",
  "err": "UserStorage.CreateUser: insertUser: db connection failed",
  "method": "CreateUser",
  "took": "7.585µs",
  "ts": "2018-12-20T13:44:52.491775Z",
  "user": {
    "ID": "",
    "Username": "alice"
  }
}
```

Errors returned from Go kit's `endpoint.Endpoint` can be propagated to the end user (requests throttling)
or shown as internal errors (JSON serialization errors, e.g., EOF):

```sh
$ curl -i -X POST -d '{"username":"bob"}' http://localhost:8000/v1/users
HTTP/1.1 429 Too Many Requests

{"error":{"code":"rate_limit","message":"API rate limit exceeded."}}
```

The error was also logged for operators:

```json
{
  "caller": "server.go:112",
  "component": "HTTP",
  "err": "rate limit exceeded",
  "ts": "2018-12-20T13:49:12.333333Z"
}
```
