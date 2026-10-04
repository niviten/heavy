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
make db-setup
go mod tidy
go run ./cmd/server
```

The Go database commands load connection settings from `.env`:

```sh
go run ./cmd/migrate # Apply sql/schema.sql
go run ./cmd/seed    # Upsert the canonical muscle groups from sql/seed.sql

# Equivalent convenience wrappers:
make migrate
make seed
make db-setup
```

The database configured by `DB_NAME` must already exist. Environment variables
set by the shell take precedence over values in `.env`.

The server exposes:

- `GET /api/json` — application information
- `GET /api/json/health` — application and MySQL health
- `GET /api/json/muscle-groups` — all muscle groups
- `GET /api/json/muscle-groups/:id` — a muscle group by ID
- `GET /api/json/muscle-groups/name/:name` — a muscle group by name
- `GET /api/json/exercises` — paginated, searchable, sortable exercises
- `GET /api/json/exercises/:id` — an exercise by ID
- `GET /api/json/muscle-groups/:muscle_group_id/exercises` — exercises in a muscle group
- `POST /api/json/exercises` — create an exercise
- `PATCH /api/json/exercises/:id` — update an exercise name and/or muscle group
- `DELETE /api/json/exercises/:id` — soft-delete an exercise
- `GET /api/json/workout-routines` — paginated, searchable, sortable workout routines
- `GET /api/json/workout-routines/:id` — a workout routine by ID
- `POST /api/json/workout-routines` — create a workout routine
- `PATCH /api/json/workout-routines/:id` — update a workout routine name and/or description
- `DELETE /api/json/workout-routines/:id` — permanently delete a workout routine

Exercise collection endpoints accept `page` (default `1`), `page_size`
(default `20`, maximum `100`), `search`, `sort_by`, and `sort_order`. Valid
sort fields are `exercise_id`, `exercise_name`, and `muscle_group_id`; sort
order is `asc` or `desc`.

Create request:

```json
{
  "exercise_name": "Incline Dumbbell Press",
  "muscle_group_id": 1
}
```

Update request:

```json
{
  "exercise_name": "Incline Press",
  "muscle_group_id": 2
}
```

Either update field can be omitted, so the same endpoint can rename an
exercise, change its muscle group, or perform both changes together.

Exercise table, column, route, and JSON field names use the corrected
`exercise` spelling. The complete schema is saved in `sql/schema.sql`.

Workout routine collection endpoints use the same pagination parameters.
Their valid sort fields are `workout_routine_id`, `workout_routine_name`, and
`workout_routine_description`. Search matches both the name and description.

Workout routine create request:

```json
{
  "workout_routine_name": "Push Day",
  "workout_routine_description": "Chest, shoulders, and triceps"
}
```

The description is optional. A workout routine update accepts the name, the
description, or both. A blank or whitespace-only description is stored as
`NULL`:

```json
{
  "workout_routine_name": "Upper Push Day",
  "workout_routine_description": "Chest, shoulders, and triceps"
}
```

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
