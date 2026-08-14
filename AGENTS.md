# AGENTS.md

## Project Overview

Lin-hub is a project encompassing a Go backend, a Vite-based frontend, Keycloak themes, and infrastructure managed via Terraform and Kubernetes.

The codebase is split into:

- **Backend API** (`/server`): Go application using Protobuf and DDD principles.
- **API CLI** (`/server/cmd/hub`): `hub`, the command line client for the API. Its command tree is generated from the protos.
- **Frontend Web** (`/ui/web`): Vite application using TypeScript and React.
- **Keycloak Theme** (`/ui/keycloak-theme`): Custom themes for Keycloak.
- **Infrastructure (Terraform)** (`/infra/tf`): IaC using Terraform.
- **Infrastructure (Kubernetes)** (`/infra/k8s`): Kubernetes manifests and overlays.
- **Documentation site** (`/docs`): the GitHub Pages site — guides, the Swagger API explorer and the generated API reference.

## Backend Workflow

- Read `server/AGENTS.md` for details.
- Use `make` commands for common tasks.
- Protobuf is used for API definitions (`/server/proto`).
- The protos are the single source of truth: `make gen` derives the gRPC stubs,
  the OpenAPI spec, the web client's operation table, the RBAC rules the server
  enforces and the API reference the `hub-api` agent skill ships. Never hand-edit
  a generated file, and never restate a REST path or a permission by hand.

## CLI Workflow

- `make cli` installs `hub`, the API client.
- `hub api list` and `hub api describe <rpc>` enumerate the API, its RBAC rules
  and the command that calls each rpc.
- `hub api call <METHOD> <PATH>` is the escape hatch for anything the generated
  commands do not cover.
- Use `--dry-run` before anything destructive.
- The `hub-api` agent skill (`.agents/skills/hub-api`) documents all of this for
  AI agents.

## Documentation Workflow

- Read `docs/AGENTS.md` for details.
- `cd docs && pnpm serve` builds and serves the site locally.
- The guides are Markdown in `docs/site/`; the API explorer and the API
  reference are assembled at build time from `server/openapi/` and the
  reference `make gen` writes. Do not copy either into the site by hand.

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
