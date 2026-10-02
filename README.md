# Heavy

Heavy is an application for tracking the workouts we do. It provides a layered
Go API built with Echo and MySQL, ready to grow with workout, exercise, set, and
progress-tracking features.

## Requirements

- Go 1.25 or newer
- MySQL 5.7+ or MySQL 8+

## Run locally

```sh
cp .env.example .env
# Update the database credentials in .env.
go mod tidy
go run ./cmd/server
```

The server exposes:

- `GET /` — application information
- `GET /health` — application and MySQL health

Run the tests with:

```sh
go test ./...
```

## Structure

```text
cmd/server             application entry point and dependency wiring
internal/config        environment configuration
internal/controller    HTTP request and response handling
internal/database      MySQL connection-pool setup
internal/dto           HTTP request and response data
internal/model         database table entities
internal/repository    database operations
internal/routes        HTTP route registration
internal/service       application logic
```
