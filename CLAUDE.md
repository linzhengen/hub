# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based microservices backend API server with a clean architecture. The project consists of:
- `server/` - Main Go backend application (gRPC/HTTP API with gRPC-Gateway)
- `ui/` - Frontend application (currently empty `web/` directory)
- `infra/` - Infrastructure as Code (Kubernetes manifests and Terraform configurations)

## Development Commands

### Server Development (in `server/` directory)

**Setup:**
```bash
cd server
make init          # Install development tools
make pre-commit-install  # Install pre-commit hooks
```

**Development:**
```bash
make dev           # Start local development with Docker Compose (MariaDB, Keycloak, Air hot-reload)
make dev-seed      # Run database seeds after code generation
make gen           # Generate code (sqlc, protobuf, OpenAPI)
make migrate       # Run database migrations
```

**Testing & Quality:**
```bash
make test          # Run all tests
make lint          # Run golangci-lint
```

**Infrastructure:**
```bash
make build         # Build container images with skaffold
make s-dev         # Start development on Kubernetes (minikube) with skaffold dev
make tunnel        # Create tunnel to minikube service
```

### Root Level Commands
```bash
make pre-commit-install  # Install pre-commit hooks for the entire repository
```

## Architecture

### Clean Architecture Structure
The server follows Clean Architecture with clear dependency direction:
- `Interface` → `Usecase` → `Domain` ← `Infrastructure`

**Key directories in `server/`:**
- `internal/domain/` - Core business entities and interfaces (no external dependencies)
- `internal/usecase/` - Application business logic
- `internal/interface/` - External interfaces (gRPC handlers, CLI)
- `internal/infrastructure/` - External implementations (database, external services)
- `pkg/` - Shared utility packages (logger, jwt, hash, uuid)
- `proto/` - Protocol Buffer definitions
- `pb/` - Generated Protobuf/gRPC code
- `db/` - Database schema and migrations
- `di/` - Dependency injection wiring (go.uber.org/dig)
- `config/` - Configuration structure

### Authentication & Authorization
- **Authentication**: Keycloak OIDC/OAuth2 integration
- **Authorization**: Custom Role-Based Access Control (RBAC) system
- **Flow**: gRPC interceptors validate JWT tokens and enforce permissions based on user group memberships

### Database
- **Primary**: MySQL/MariaDB
- **Migrations**: golang-migrate (`db/migrations/mysql/`)
- **Queries**: sqlc for type-safe SQL queries (`internal/infrastructure/persistence/mysql/query/`)
- **Transactions**: Managed via `trans` package

### API Design
- **Primary**: gRPC with Protocol Buffers
- **REST**: gRPC-Gateway provides RESTful JSON API on same port
- **Routing**: cmux distinguishes between gRPC and HTTP traffic
- **Code Generation**: buf for Protobuf linting and code generation

## Development Workflow

For detailed development workflow, refer to `server/CLAUDE.md` which contains comprehensive guidelines including:

1. **Ticket Understanding** - Check `GEMINI.md` for architecture conventions
2. **Branch Creation** - Follow commitlint conventions (`type/short-description`)
3. **Domain Definition** - Add to `internal/domain/`
4. **Protobuf Definition** - Add to `proto/`
5. **SQL Definition** - Create queries for sqlc
6. **Code Generation** - Run `make gen`
7. **Repository Implementation** - Implement in `internal/infrastructure/`
8. **Usecase Implementation** - Implement in `internal/usecase/`
9. **gRPC Handler Implementation** - Implement in `internal/interface/grpc/`
10. **Dependency Injection** - Add to `di/di.go`
11. **Static Analysis** - Run `golangci-lint run ./...`
12. **Testing** - Create and run tests with `go test ./...`
13. **Commit & Push** - Follow commitlint conventions
14. **Pull Request Creation** - Create PR on GitHub

## Code Generation

The project relies heavily on code generation:
- **`make gen`** runs:
  - `sqlc generate` - Type-safe SQL queries
  - `buf generate` - Protobuf/gRPC code
  - OpenAPI spec patching
  - Protobuf to YAML conversion

## Git & Commit Conventions

- **Commit Messages**: Follow Conventional Commits specification (enforced by commitlint)
- **Branch Names**: `type/short-description` (e.g., `feat/add-user-api`, `fix/login-issue`)
- **Pre-commit Hooks**: Installed via `make pre-commit-install`, includes:
  - commitlint validation
  - trailing whitespace removal
  - golangci-lint
  - buf linting/formatting

## Infrastructure

- **Kubernetes**: Manifests in `infra/k8s/` (base and overlays)
- **Terraform**: Configurations in `infra/tf/` (dev environment, Keycloak, modules)
- **Local Development**: Docker Compose with MariaDB and Keycloak

## Testing

- **Framework**: Standard Go testing with testify assertions
- **Mocking**: go-sqlmock for database testing
- **Coverage**: Tests exist at all layers (domain, usecase, infrastructure, interfaces)
- **Run**: `make test` or `go test ./...`

## Key Technologies

- **Go 1.26+** with go-zero (microservices), go.uber.org/dig (DI)
- **Database**: MySQL, golang-migrate, sqlc
- **API**: gRPC, gRPC-Gateway, Protocol Buffers, OpenAPI
- **Auth**: Keycloak, JWT, custom RBAC
- **Infrastructure**: Docker, Kubernetes, Terraform
- **Development**: Air (hot-reload), pre-commit hooks, buf

## Important Notes

- The server already has a comprehensive `server/CLAUDE.md` with detailed implementation guidelines
- Follow the Clean Architecture dependency direction strictly
- Use dependency injection via `di/di.go` for all components
- Always run `make gen` after modifying Protobuf definitions or SQL queries
- Database changes require migrations in `db/migrations/mysql/`
- Authentication/authorization is handled via gRPC interceptors
- Use CQRS pattern for complex queries (separate Finder interfaces)
