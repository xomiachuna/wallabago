# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wallabago is a Golang implementation of Wallabag (read-it-later service). It's a web service that allows users to save web articles for later reading, with OAuth2 authentication and role-based access control.

## Development Commands

### Running the Application
```bash
make up      # Start with docker-compose (includes build, tidy, format, codegen)
make down    # Stop docker-compose
```

### Testing
```bash
make test                # Run all tests
go test -v ./...         # Run all tests with verbose output
go test -v ./test        # Run BDD tests specifically
dlv test ./test          # Interactive debugging with Delve (for TDD/BDD)
```

### Code Quality
```bash
make check               # Run full check (format, lint, tests, diagrams, tidy)
make check-quick         # Quick check (no tests)
make lint                # Run golangci-lint
make format              # Format code with golangci-lint
make tidy                # Run go mod tidy
```

### Code Generation
```bash
make codegen             # Generate all code (sqlc)
make sqlc                # Generate database code from SQL (via sqlc)
```


### Documentation
```bash
make diagrams            # Generate PlantUML diagrams
make adr                 # Generate ADR documentation
./tools/adr.sh           # Interact with Architecture Decision Records
```

## Architecture

### Layered Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

1. **`cmd/wallabago-api`**: Application entry point. Reads environment variables and initializes the server.

2. **`internal/app`**: Application bootstrap layer. The `Wallabago` struct orchestrates all managers, engines, and infrastructure (database, instrumentation). The `Handler()` method wires up HTTP routes with middleware.

3. **`internal/managers`**: Business logic orchestration layer. Managers coordinate between engines and storage, handle transactions, and enforce authorization policies. Examples:
   - `IdentityManager`: OAuth2 token management and user authentication
   - `EntryManager`: Entry creation, retrieval, and authorization
   - `BootstrapManager`: Initial system setup

4. **`internal/engines`**: Specialized business logic implementations:
   - `RBACAuthorizationEngine`: Role-based access control enforcement
   - `SimpleReadabilityRetrievalEngine`: Article content extraction from URLs
   - `BoostrapEngine`: Initial system setup logic

5. **`internal/core`**: Domain models and business rules. Contains pure domain logic without dependencies on infrastructure. Includes the `policy` package for RBAC definitions.

6. **`internal/storage`**: Data persistence layer. Wraps the `database` package and provides higher-level storage abstractions.

7. **`internal/database`**: Generated database access code (sqlc). Contains migrations, queries, and generated Go code for type-safe SQL operations.

8. **`internal/http`**: HTTP server and handlers. Contains middleware (auth, logging, panic recovery, OpenTelemetry) and API handlers.

9. **`internal/instrumentation`**: Observability setup (OpenTelemetry).

### Key Patterns

- **Dependency Injection**: The `Wallabago` struct in `internal/app/wallabago.go` constructs the entire dependency graph. Managers receive their dependencies (engines, storage) as constructor parameters.

- **Interface-based Design**: Managers depend on interfaces (defined in `internal/managers/dependencies.go`) rather than concrete implementations. This enables testing and flexibility.

- **Transaction Management**: Authorization checks and database operations are wrapped in transactions. The `tx` parameter flows through manager methods to engines.

- **Error Handling**: Uses `github.com/pkg/errors` for error wrapping with stack traces. User-facing errors are distinguished from internal errors (see `internal/core/errors.go`).

### OAuth2 Flow

Authentication uses OAuth2 with password grant type. Bootstrap creates a default admin user and client. Access tokens are validated via the `OAuth2Middleware` in `internal/http/middleware/auth.go`.

### RBAC Authorization

**Note**: Authorization is currently a non-functional stub (allows everything). The `RBACAuthorizationEngine` exists but is not fully implemented. This is intentional and will be revisited later.

Access control is designed to use a role-based system with permissions defined in `internal/core/policy`. Managers call `authz.CheckPolicy()` to verify permissions before operations. See `docs/ACCESSCONTROL.md` for the intended design.

### BDD Testing

The project uses Cucumber/Godog for BDD tests. Test scenarios are in `features/*.feature` files. The test suite (`test/bdd_test.go`) uses temporary SQLite databases and runs scenarios against a real server instance.

### Database and Code Generation

- Database schema is defined in SQL migrations (`internal/database/migrations/`)
- Queries are in `internal/database/queries.sql`
- sqlc generates type-safe Go code from SQL (configured in `sqlc.yaml`)
- Always run `make codegen` after modifying SQL files or migrations

## Environment Variables

- `WALLABAGO_PORT`: Server port (default: 8080)
- `WALLABAGO_DB_PATH`: SQLite database file path (default: ./wallabago.db)
- `WALLABAGO_BOOTSTRAP_ADMIN_USERNAME`: Bootstrap admin username (default: admin)
- `WALLABAGO_BOOTSTRAP_ADMIN_PASSWORD`: Bootstrap admin password (default: admin)
- `WALLABAGO_BOOTSTRAP_ADMIN_EMAIL`: Bootstrap admin email (default: admin@admin.co)
- `WALLABAGO_BOOTSTRAP_CLIENT_ID`: Bootstrap client ID (default: web)
- `WALLABAGO_BOOTSTRAP_CLIENT_SECRET`: Bootstrap client secret (default: web)
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: Enable OpenTelemetry tracing

## Pre-commit Hooks

The project uses pre-commit hooks (`.pre-commit-config.yaml`). Install with:
```bash
pre-commit install
```

## Documentation

- Architecture diagrams: `docs/ARCHITECTURE.md` (source: `docs/diagrams/*.puml`)
- Access control design: `docs/ACCESSCONTROL.md`
- Architecture Decision Records: `docs/adr/`
- Use cases: `docs/USECASES.md`

## Development Workflow

- Do not prompt for or execute git commands (add, commit, push, etc.) unless the user specifically requests them
- just parse it on handler layer, keep the core typesafe
- use generated always as identity instead of serial