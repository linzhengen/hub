# AGENTS.md

## Project Overview

Lin-hub is a project encompassing a Go backend, a Vite-based frontend, Keycloak themes, and infrastructure managed via Terraform and Kubernetes.

The codebase is split into:

- **Backend API** (`/server`): Go application using Protobuf and DDD principles.
- **Frontend Web** (`/ui/web`): Vite application using TypeScript and React.
- **Keycloak Theme** (`/ui/keycloak-theme`): Custom themes for Keycloak.
- **Infrastructure (Terraform)** (`/infra/tf`): IaC using Terraform.
- **Infrastructure (Kubernetes)** (`/infra/k8s`): Kubernetes manifests and overlays.

## Backend Workflow

- Read `server/AGENTS.md` for details.
- Use `make` commands for common tasks.
- Protobuf is used for API definitions (`/server/proto`).

## Frontend Workflow

- Read `ui/web/AGENTS.md` for details.
- Use `pnpm` for package management.

## Infrastructure Workflow

- Read `infra/tf/AGENTS.md` for Terraform details.
- Read `infra/k8s/AGENTS.md` for Kubernetes details.

## UI Components

- Read `ui/keycloak-theme/AGENTS.md` for Keycloak theme details.

## Testing & Quality Practices

- Follow TDD: red → green → refactor.
- Enforce strong typing; avoid `any` or `interface{}` where possible.
- Write self-documenting code; only add comments that explain intent.

## Language Style

- **Go**: Follow standard Go idioms. Use `golangci-lint`.
- **TypeScript**: Use strict mode, rely on ESLint, and avoid `any` types.

## General Practices

- Prefer editing existing files; add new documentation only when requested.
- Inject dependencies through constructors and preserve clean architecture boundaries.
- Handle errors with domain-specific exceptions/errors at the correct layer.
