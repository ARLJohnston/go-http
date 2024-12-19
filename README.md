[![codecov](https://codecov.io/gh/ARLJohnston/go-http/graph/badge.svg?token=LA0NE8ENYZ)](https://codecov.io/gh/ARLJohnston/go-http)

# Example microservice application

This repository is an example microservice architecture application

[![](https://mermaid.ink/img/pako:eNpdUctugzAQ_JWVz4l66wFVlSABIjVt0ySnQg8LmGAVbMuYtlGSf68f5NGemJ0Z74zNgZSioiQgULfiu2xQaViucw79UOwUygZCKVtWomaCGxqgYoqWdvI-gDBLlOB6GvMKUnG3pZ1sP5wSZRGWn6PgqVk2R40F9hSe95u3pWH9EniYTh-Pu_VqdoTIn3YUzEaH-1Be3XbbKqxrVjr9tlm0HSmAefZ0D3YtaG_-uEixlRqt5R_JR5iDYNN9ldjh0De5pL8WPVVfWLCW6b1V_j9Nkq2U6Khu6NC73Yvsxbw2xD9SKE2V49IsNenI0U4-IXRxiYXRFSYOpt6yuPKLc7kzjMZrkAnpqOqQVeb_HiyZE9OlozkJDKxojUOrc5Lzk7HioMVmz0sSaDXQCVFi2DUkqLHtzTTICjWdMzRX784WifxdiG40nX4BlhS4Rg?type=png)](https://mermaid.live/edit#pako:eNpdUctugzAQ_JWVz4l66wFVlSABIjVt0ySnQg8LmGAVbMuYtlGSf68f5NGemJ0Z74zNgZSioiQgULfiu2xQaViucw79UOwUygZCKVtWomaCGxqgYoqWdvI-gDBLlOB6GvMKUnG3pZ1sP5wSZRGWn6PgqVk2R40F9hSe95u3pWH9EniYTh-Pu_VqdoTIn3YUzEaH-1Be3XbbKqxrVjr9tlm0HSmAefZ0D3YtaG_-uEixlRqt5R_JR5iDYNN9ldjh0De5pL8WPVVfWLCW6b1V_j9Nkq2U6Khu6NC73Yvsxbw2xD9SKE2V49IsNenI0U4-IXRxiYXRFSYOpt6yuPKLc7kzjMZrkAnpqOqQVeb_HiyZE9OlozkJDKxojUOrc5Lzk7HioMVmz0sSaDXQCVFi2DUkqLHtzTTICjWdMzRX784WifxdiG40nX4BlhS4Rg)

# Development
To enter into an environment with all necessary tooling installed:
```console
nix develop
```
This will enter an environment with all the packages listed in the mkShell in  the [nix flake](./flake.nix)

# Deployment
## Individual services
### Database
Start a MySQL database on port 3306, and run the [sql migration script](./server/create-tables.sql)

This can be done with Docker as follows:
```console
cd server
docker compose up -d
```

To remove the database and all data in it:
```console
docker compose down -v
```

### Database Client
Requires a running MySQL database, which is set with `MYSQL_DATABASE_ADDRESS` and defaults to `localhost:3306`
Use Go to build and run the client:
```console
cd server
go run main.go
```
This will start a gRPC server on :50051 and a http server on port :2121

### Front-end
Requires [templ](https://github.com/a-h/templ/) to be installed.
Use Go to build and run the front end:
```console
cd front
templ generate
go run .
```
This will start a http server on port :3000

## All-in-one
Deployment scripts are in the [deployments directory](./deployments/) and contains instructions on running the application with: [Docker Compose](https://docs.docker.com/compose/), [Docker Swarm](https://docs.docker.com/engine/swarm/) and [Kubernetes](https://kubernetes.io/) (with and without service mesh frameworks).
