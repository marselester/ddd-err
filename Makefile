test:
	go test ./... -v

lint:
	golangci-lint run

TEST_PGPORT := 5436
TEST_PGDATABASE := test_account
TEST_PGUSER := test_account
TEST_PGPASSWORD := swordfish
export TEST_PGPORT TEST_PGDATABASE TEST_PGUSER TEST_PGPASSWORD

docker_run_postgres:
	docker run --rm -p 5436:5432 -e POSTGRES_USER=$(TEST_PGUSER) -e POSTGRES_PASSWORD=$(TEST_PGPASSWORD) postgres:12-alpine
