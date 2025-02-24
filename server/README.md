[![Go Report Card](https://goreportcard.com/badge/github.com/ARLJohnston/go-http/server)](https://goreportcard.com/report/github.com/ARLJohnston/go-http/server)

# Running the client
Requires a running Postgres database, which is set with `DATABASE_ADDRESS` and defaults to `localhost:5432`

Use Go to build and run the client:
```console
go run main.go
```
This will start a gRPC server on :50051 and a http server on port :2121

## Running the database with Docker
A [docker compose file](./docker-compose.yml) is provided for running a database
```console
docker compose up -d
```

To remove the database and all data in it:
```console
docker compose down -v
```

# Test
To run the tests
```console
go test .
```
