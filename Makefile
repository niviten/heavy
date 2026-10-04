GO ?= go

.PHONY: migrate seed db-setup

migrate:
	$(GO) run ./cmd/migrate

seed:
	$(GO) run ./cmd/seed

db-setup: migrate seed
