# Getting started

This page takes you from a fresh clone to a running stack, and then through the
day-to-day loop of changing the API, the web app or the infrastructure.

## Prerequisites

| Tool | Used for |
|:---|:---|
| Docker + Docker Compose | PostgreSQL, Keycloak, MailHog and the API server |
| Go 1.26+ | the backend, the code generators and the `hub` CLI |
| Node.js + pnpm | the React web app and this documentation site |
| Terraform | seeding the Keycloak realm and clients |
| `kubectl` + `minikube` | the optional Kubernetes deployment |

## 1. Install the toolchain

```bash
git clone https://github.com/linzhengen/hub.git
cd hub

# Backend generators and hooks: buf plugins, sqlc, migrate, pre-commit
make init

# Frontend packages
cd ui/web && pnpm install && cd ../..
```

## 2. Start the local stack

```bash
make dev
```

`make dev` brings up Docker Compose, waits for Keycloak to become healthy,
relaxes the master realm's TLS requirement for local use, applies the Terraform
in `infra/tf/dev` to create the realm and clients, seeds the database, and then
tails the logs.

| Service | URL | Notes |
|:---|:---|:---|
| API (gRPC-Gateway) | <http://localhost:9090> | REST surface under `/api/v1` |
| Keycloak | <http://localhost:8080> | admin console, `admin` / `admin` |
| PostgreSQL | `localhost:5432` | database `hub` |
| MailHog | <http://localhost:8025> | catches verification mail |
| Web UI | <http://localhost:3000> | started separately, see below |

The web app runs outside Compose, against the Vite dev server:

```bash
cd ui/web && pnpm dev
```

## 3. Seed data

`make dev` already seeds, but you can re-run it at any time:

```bash
make dev-seed
```

That runs `seed` (the initial users, groups, roles and permissions) and
`resource-import` (the menu/resource tree the UI renders) inside the API
container.

## 4. Talk to the API

The quickest check that everything is wired up is the CLI:

```bash
make cli   # installs the `hub` binary

HUB_ENDPOINT=http://localhost:9090 \
HUB_OIDC_ISSUER=http://localhost:8080/realms/hub \
HUB_OIDC_CLIENT_ID=hub-web \
hub auth login

hub auth whoami
hub user list -o table
```

The [CLI guide](cli.html) covers profiles, the other grant types and the whole
command surface.

## Project layout

```text
.
├── server/                 # Go backend
│   ├── cmd/                # entry points: server, hub CLI, generators
│   ├── internal/           # domain / usecase / interface / infrastructure
│   ├── proto/              # API definitions — the single source of truth
│   ├── openapi/            # generated OpenAPI documents
│   └── pkg/hubcli/         # the CLI, built from the API catalog
├── ui/
│   ├── web/                # React SPA
│   └── keycloak-theme/     # custom Keycloak login theme
├── infra/
│   ├── tf/                 # Terraform (Keycloak realm, clients)
│   └── k8s/                # Kubernetes manifests (Kustomize base + overlays)
├── docs/                   # this documentation site
└── Makefile                # common tasks
```

Every top-level directory carries an `AGENTS.md` with the conventions for that
part of the codebase. Read it before changing code there.

## The generation loop

The `.proto` files are the single source of truth. After changing one — or a
sqlc query — run:

```bash
make gen
```

That regenerates, in order: sqlc code, the DI adapters, the gRPC and
gRPC-Gateway stubs, the OpenAPI documents under `server/openapi/`, the web
client's operation table (`ui/web/src/api/operations.ts`), the API reference the
agent skill ships, and the web app's TypeScript schema types.

**Never hand-edit a generated file, and never write a REST path or a permission
twice.** CI regenerates the Go-derived artifacts and fails if the checked-in
copies are stale.

A typical backend change follows the workflow in
[`server/AGENTS.md`](https://github.com/linzhengen/hub/blob/main/server/AGENTS.md):
domain → proto → SQL → `make gen` → repository → usecase → gRPC handler → DI →
lint → test.

## Checks

```bash
make lint                     # golangci-lint over the server
make test                     # go test ./...
cd ui/web && pnpm lint        # tsc --noEmit
cd ui/web && pnpm lint:eslint
cd ui/web && pnpm test        # vitest
cd ui/web && pnpm test:e2e    # Playwright
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/); the
`commit-msg` hook installed by `make init` enforces it.

## Database migrations

Migrations live in `server/db/migrations/postgres` and run with
`golang-migrate`:

```bash
make migrate
```

## Deploy to Kubernetes (minikube)

```bash
# Build the image inside minikube's Docker daemon
eval $(minikube docker-env) && docker build -t hub .

# Preview, then apply
kubectl kustomize infra/k8s/overlays/minikube
kubectl apply -k infra/k8s/overlays/minikube
```

`infra/k8s/base` is written for the disposable dev and minikube overlays: it
enables plain HTTP on Keycloak and runs PostgreSQL without persistence. Read
[`infra/k8s/AGENTS.md`](https://github.com/linzhengen/hub/blob/main/infra/k8s/AGENTS.md)
before reusing it anywhere else — it lists exactly what an overlay must patch
first.

## Documentation site

This site is built from `docs/` and published to GitHub Pages by
`.github/workflows/pages.yml` on every push to `main`. To work on it locally:

```bash
cd docs
pnpm install
pnpm serve      # builds docs/_site and serves it on http://localhost:4000
```

The guides are Markdown under `docs/site/`. The API explorer and the API
reference are not written by hand — they are assembled at build time from
`server/openapi/` and from the reference `make gen` produces, so they cannot
drift from the protos.
