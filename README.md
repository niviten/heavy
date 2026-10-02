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

- `GET /api/json` — application information
- `GET /api/json/health` — application and MySQL health
- `GET /api/json/muscle-groups` — all muscle groups
- `GET /api/json/muscle-groups/:id` — a muscle group by ID
- `GET /api/json/muscle-groups/name/:name` — a muscle group by name

Run the tests with:

```sh
go test ./...
gotestsum ./...
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
