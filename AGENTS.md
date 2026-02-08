# Agent Guidelines for movieswithfriends

This document provides essential information for AI coding agents working in this repository.

## Project Overview

**Tech Stack:** Go 1.23, PostgreSQL 17, htmx, Bootstrap 5, OpenTelemetry  
**Architecture:** Domain-Driven Design with clear bounded contexts  
**Domains:** `identityaccess` (users/auth), `partymgmt` (parties/movies)

## Build, Test, and Lint Commands

### Development
```bash
make run                    # Start app with docker-compose + hot reload
make build-assets           # Build & minify JavaScript assets with fingerprinting
make migrate                # Run database migrations (goose)
make seed                   # Seed database with test data (use -drop to reset)
make psql                   # Connect to PostgreSQL via docker
```

### Testing
```bash
go test ./...               # Run all Go unit tests
go test ./path/to/package   # Run tests for specific package
go test -run TestName       # Run single test by name
make e2e                    # Run all Playwright e2e tests
make e2e TEST=testname      # Run single e2e test (e.g., TEST=login)
make e2e headless=false     # Run e2e tests with visible browser
```

### Linting
```bash
make lint                   # Run vet + staticcheck
make vet                    # Run go vet only
make staticcheck            # Run staticcheck only
```

### Infrastructure
```bash
make pkr                    # Build Packer AMI
make tf-plan                # Terraform plan
make tf-apply               # Terraform apply
make deploy                 # Deploy with Kamal
```

## Code Style Guidelines

### Go Conventions

#### Imports
- Group imports: stdlib → third-party → local
- Separate groups with blank lines
- Use goimports for automatic formatting

```go
import (
    "context"
    "errors"
    "fmt"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "go.opentelemetry.io/otel"
    
    "github.com/jm96441n/movieswithfriends/identityaccess"
    "github.com/jm96441n/movieswithfriends/metrics"
)
```

#### Naming Conventions
- **Packages:** Lowercase, single word (`partymgmt`, `identityaccess`)
- **Files:** Snake_case (`party_repository.go`, `tmdb_client.go`)
- **Structs/Interfaces:** PascalCase (`PartyService`, `UserRepository`)
- **Functions/Methods:** PascalCase for exported, camelCase for unexported
- **Variables:** camelCase
- **Constants:** camelCase or UPPER_CASE for package-level

#### Error Handling
- Define custom error variables at package level:
  ```go
  var ErrPartyNotFound = errors.New("party not found")
  ```
- Wrap errors with context:
  ```go
  return fmt.Errorf("failed to create party: %w", err)
  ```
- Return errors as last return value
- Use custom error types for validation errors with multiple fields

#### Context Usage
- Always pass `context.Context` as first parameter
- Propagate context through all layers (handlers → services → repositories)
- Use context for cancellation, timeouts, and request-scoped values

#### Testing Patterns
- Use `t.Parallel()` for parallel test execution
- Table-driven tests with map[string]func pattern:
  ```go
  tests := map[string]func(t *testing.T){
      "success case": func(t *testing.T) { /* ... */ },
      "error case": func(t *testing.T) { /* ... */ },
  }
  ```
- Use testhelpers for common setup (database, user creation)
- Clean up with `testhelpers.CleanupAndResetDB(t, db)`
- E2E tests use Playwright with page object pattern

### Database Patterns

#### Repository Layer
- Use repository pattern for data access (`*Repository` structs)
- Interface type `Querier` abstracts pgx.Tx and pgxpool.Pool
- Transaction wrapper pattern for atomic operations
- Callback pattern for row scanning:
  ```go
  func (r *Repository) GetItems(ctx context.Context, fn func(item Item)) error
  ```

#### Queries
- Write raw SQL (no ORM)
- Use parameter placeholders: `$1`, `$2`, etc.
- Migrations live in `migrations/migrations/` (Goose format)
- Test isolation: schema-per-test with Testcontainers

### Observability

#### Telemetry
- Create spans for all significant operations:
  ```go
  ctx, span := metrics.SpanFromContext(ctx, "operation.name")
  defer span.End()
  ```
- Add attributes to spans for debugging:
  ```go
  span.SetAttributes(attribute.String("user.id", userID))
  ```
- Use structured logging with `slog`
- Instrument HTTP handlers with telemetry middleware

### Web Layer

#### HTTP Handlers
- Handlers return status code and error:
  ```go
  func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) (int, error)
  ```
- Use middleware for auth, telemetry, recovery
- Parse forms with `r.ParseForm()`
- Redirect with `http.Redirect(w, r, "/path", http.StatusSeeOther)`

#### Templates
- Use Go html/template with `.gohtml` extension
- Base template pattern with blocks: `{{ define "base" }}`
- Template functions registered (e.g., `assetPath` for fingerprinted assets)
- Flash messages for user feedback via sessions

#### Frontend
- Vanilla JavaScript (no frameworks)
- htmx for AJAX interactions (v1.9.9)
- Bootstrap 5.3.2 for UI components
- Event-driven architecture with custom events
- Asset pipeline: custom Go tool for minification + fingerprinting

### Configuration

- All config via environment variables (use `.envrc` for local dev)
- No hardcoded secrets in code
- Docker Compose for local development
- Multi-stage Dockerfile (dev and prod targets)

## Architecture Principles

1. **Domain-Driven Design:** Keep domain logic in domain packages
2. **Repository Pattern:** Data access abstraction in `store/` directories
3. **Service Layer:** Business logic in `*Service` structs
4. **No ORM:** Direct SQL with pgx driver for clarity and performance
5. **Embedded Assets:** Use `//go:embed` for templates and static files
6. **Test Isolation:** Schema-per-test pattern prevents test interference
7. **Observability-First:** Instrument everything for production debugging

## Key Files and Directories

```
cmd/movieswithfriends/      # Application entry point
config/                     # Configuration loading
identityaccess/             # User authentication domain
  └── store/                # User repository
partymgmt/                  # Party management domain
  └── store/                # Party repository
migrations/migrations/      # Database migrations (22 files)
web/                        # HTTP handlers and middleware
ui/html/                    # Go templates
ui/static/js/               # JavaScript source files
ui/dist/                    # Built/fingerprinted assets
tools/                      # Dev tools (asset builder, keygen, seed)
testhelpers/                # Shared test utilities
e2e/                        # Playwright e2e tests (11 tests)
infra/                      # Infrastructure as code
```

## Common Tasks

### Adding a New Feature
1. Identify the domain (identityaccess or partymgmt)
2. Add domain logic to domain package
3. Create/update repository methods in `store/`
4. Add service methods if needed
5. Create HTTP handler in `web/`
6. Add template in `ui/html/`
7. Write tests at each layer
8. Add e2e test if user-facing

### Database Changes
1. Create migration: `goose -dir migrations/migrations create migration_name sql`
2. Write up migration SQL
3. Write down migration SQL
4. Run migrations: `make migrate`
5. Update repository layer to use new schema

### Adding Dependencies
```bash
go get github.com/package/name
go mod tidy
```

## Notes for Agents

- Always run tests after changes: `go test ./...`
- Use `make lint` before committing
- Format Go templates with Prettier: `make setupprettier` first
- Database tests require Docker (Testcontainers)
- E2E tests require Playwright: see `e2e/README.md`
- Asset changes require rebuild: `make build-assets`
