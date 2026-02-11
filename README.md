# Aogeri API

This repository implements the backend API for Aogeri — a staking, liquidity and governance platform. The README below documents the tech stack, architecture, how to run the project, and a short set of user stories with acceptance test cases you can use to validate the app behavior.

## Tech stack

- Language: Go (1.21)
- Web router: chi
- DB: PostgreSQL (pgx / pgtype types used by sqlc)
- SQL code generation: sqlc
- Containerization: Docker / Docker Compose
- Redis: for token/session store
- Linting: golangci-lint

## High-level architecture

- HTTP API (cmd/api) registers handlers that call services which invoke generated `db.Queries` (sqlc).
- Services contain business logic (staking, governance, assets). Handlers convert HTTP requests to service calls and format responses.
- Database uses typed pgx `pgtype` structures for JSON mapping and to avoid low-level conversions in SQL code.
- Authentication: JWTs, stored refresh tokens in Redis via a store adapter.

## Getting started (local / development)

Prerequisites:
- Docker & docker-compose
- Go toolchain (for local build/test; optional if you run via Docker)

Run the full stack with Docker Compose (recommended):

```bash
# build and start services (api, postgres, redis)
make docker-up
```

Check the API health:

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health
# should return 200
```

Notes about database migrations:
- The repository includes SQL migration files in `internal/db/migrations`.
- The `Makefile` has `migrate-up` target that calls `goose`. If you don't have `goose` installed, apply migrations manually inside the running Postgres container:

```bash
# from repo root
docker compose exec -T postgres psql -U postgres -d aogeri -f internal/db/migrations/000001_init_schema.up.sql
docker compose exec -T postgres psql -U postgres -d aogeri -f internal/db/migrations/000003_seed_ui_upsert.sql
```

There is also a seed SQL file used during our session to insert sample tokens, sample stakes, liquidity pool, security monitors and governance proposals: `internal/db/migrations/000003_seed_ui_upsert.sql`.

## Useful Make targets

- `make build` — build binary into `bin/`
- `make run` — build and run local binary
- `make dev` — run `go run ./cmd/api` (development mode)
- `make docker-up` — builds and starts containers with `docker compose`
- `make docker-down` — stops containers
- `make fmt` — run `gofmt -w .`

## Important files and packages

- `cmd/api` — API entrypoint and router setup
- `internal/handlers` — HTTP handlers
- `internal/services` — business logic
- `internal/db` — generated SQLC code and migrations
- `internal/auth` — authentication logic and middleware
- `internal/models` — internal request/response types
- `tester.app.http` — collection of example HTTP requests you can run against a running server

## Main API endpoints (quick reference)

- POST /api/v1/register — register new user
- POST /api/v1/login — login (returns access_token, refresh_token)
- POST /api/v1/refresh — refresh tokens
- POST /api/v1/logout — logout (invalidate refresh token)
- GET /api/v1/stakes — list user stakes (authenticated)
- POST /api/v1/stakes — create stake (authenticated)
- POST /api/v1/stakes/{id}/unstake — unstake a stake (authenticated)
- GET /api/v1/assets — list assets
- GET /api/v1/proposals — list governance proposals
- GET /health — health check

See `tester.app.http` for copy-paste ready requests and examples.

## User stories and test cases

Below are a few example user stories described in plain English and paired with simple acceptance test steps (Given/When/Then) so you or QA can verify correct behavior.

1) User registration and login

    - Story: As a user I want to create an account and obtain credentials so I can interact with protected endpoints.
    - Acceptance test:
      - Given the API is running
      - When I POST /api/v1/register with email/password conforming to validation rules
      - Then the API responds 201 with a created user id
      - When I POST /api/v1/login with the same credentials
      - Then the API responds 200 with `access_token` and `refresh_token` strings

2) Create a stake

    - Story: As an authenticated user I can create a staking position for a supported token.
    - Acceptance test:
      - Given I have a valid `access_token` (from login)
      - When I POST /api/v1/stakes with body { token_symbol: "AOG", amount: "10.5", duration_days: 30 }
      - Then the API responds 201 and returns the created stake object with fields: id, user_id, token_symbol, amount, apy, start_date, end_date, status

3) View my stakes

    - Story: As an authenticated user I can list my active stakes.
    - Acceptance test:
      - Given I created a stake
      - When I GET /api/v1/stakes with Authorization: Bearer <access_token>
      - Then the API returns 200 and a JSON array containing the stake

4) Unstake flow

    - Story: As an authenticated user I can unstake a stake I own.
    - Acceptance test:
      - Given I created a stake and have its id
      - When I POST /api/v1/stakes/{id}/unstake with Authorization: Bearer <access_token>
      - Then the API responds 200 and stake status becomes `unstaked` in the database

5) Dashboard & assets (read-only)

    - Story: As a user I want to see dashboard stats and available assets.
    - Acceptance test:
      - Given the DB was seeded with assets and stakes
      - When I GET /api/v1/dashboard/stats (authenticated)
      - Then I receive JSON containing expected aggregated fields (total TVL, active stakes, etc.)
      - When I GET /api/v1/assets
      - Then I receive a list of active tokens and associated asset info

## Running tests

The project uses Go tests for unit-level checks. Run all tests with:

```bash
gofmt -w .
go test ./... -v
```

## Troubleshooting

- If `make migrate-up` fails with `goose: command not found`, install `goose` or apply migration SQL directly in the Postgres container (see migrations section).
- If Docker build fails during `go build` step, run `gofmt -w . && go build ./...` locally to surface compile errors, fix them, then rebuild.

## Next steps / To-do (suggested)

- Improve and expand service-level tests (staking calculations, rewards)
- Add automated migration runner using a well-known tool in the container image
- Add an admin/scripted seeder for deterministic test data

---

If you'd like, I can:

- generate a `scripts/seed.sh` that applies migrations and seeding SQL in the running container,
- add a short integration test that runs against a disposable Postgres container, or
- produce the Postman/Insomnia collection from `tester.app.http`.

Tell me which of the above you'd like next.
