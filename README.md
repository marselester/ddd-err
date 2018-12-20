# DDD Error Handling

[![Documentation](https://godoc.org/github.com/marselester/ddd-err?status.svg)](https://godoc.org/github.com/marselester/ddd-err)
[![Go Report Card](https://goreportcard.com/badge/github.com/marselester/ddd-err)](https://goreportcard.com/report/github.com/marselester/ddd-err)

This small project is an error handling example based on Ben Johnson's article.
Based on https://middlemost.com/failure-is-your-domain/ error can be:

- well-defined errors. These allow us to manage our application flow
  because we know what to expect and can work with them on a case-by-case basis.
- undefined errors. It can also occur when APIs we depend on add additional
  errors conditions after we've integrated our code with them.

Error consumers:

- app itself can recover from error states using error codes.
- end user requires human-readable message. API undefined errors must not be
  shown, e.g., Postgres error can reveal db schema.
- operator should be able to debug and see all errors including stack trace.

API errors should have `Code` and `Message`. For instance, "duplicate username" error

```go
&account.Error{
	Code:    account.EConflict,
	Message: "Username is already in use. Please choose a different username.",
}
```

is shown to the API consumer

```
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

```
$ curl -i -X POST -d '{"username":"alice"}' http://localhost:8000/v1/users
HTTP/1.1 500 Internal Server Error
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

Errors returned from a Go kit Endpoint can be propagated to the end user (requests throttling)
or shown as internal errors (JSON serialization errors, e.g., EOF):

```
$ curl -i -X POST -d '{"username":"bob"}' http://localhost:8000/v1/users
HTTP/1.1 429 Too Many Requests
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
