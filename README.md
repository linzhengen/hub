# hub

[日本語版 (Japanese)](README.ja.md) | [简体中文 (Chinese)](README.zh.md)

hub is a project that integrates Go backend, Vite frontend, Keycloak theme, and infrastructure management using Terraform and Kubernetes.

## Project Overview

This project consists of a robust backend based on Clean Architecture, a frontend using a modern tech stack, and Infrastructure as Code (IaC) to support them.

### Key Features
- **Authentication & Authorization**: Secure authentication base using Keycloak.
- **API Foundation**: RESTful API using gRPC and gRPC-Gateway.
- **UI**: Responsive management screen using React 19 and Tailwind CSS 4.
- **Infrastructure**: Resource management with Terraform and deployment to Kubernetes.

---

## Architecture

The project consists of the following four main components.

```mermaid
graph TD
    Client[Web Browser] -->|HTTPS| Gateway[gRPC-Gateway / REST API]
    Gateway -->|gRPC| Server[Go Backend API]
    Server -->|SQL| DB[(PostgreSQL)]
    Server -->|OIDC/SAML| Auth[Keycloak]
    Client -->|OIDC| Auth
    Infra[Terraform / K8s] -.->|Manages| Server
    Infra -.->|Manages| Auth
    Infra -.->|Manages| DB
```

### Layer Structure
- **Backend API (`/server`)**: Implementation of DDD (Domain-Driven Design) and Clean Architecture in Go.
- **Frontend Web (`/ui/web`)**: SPA using Vite, React 19, Tailwind CSS 4, and TanStack Query.
- **Keycloak Theme (`/ui/keycloak-theme`)**: Custom login theme for Keycloak.
- **Infrastructure (`/infra`)**:
    - `tf/`: Cloud and middleware configuration using Terraform.
    - `k8s/`: Kubernetes manifests and Kustomize overlays.

---

## Tech Stack

| Layer | Technology / Tool |
| :--- | :--- |
| **Backend** | Go 1.25, gRPC, gRPC-Gateway, Protocol Buffers, sqlc, golangci-lint |
| **Frontend** | React 19, Vite, TypeScript, Tailwind CSS 4, Shadcn UI, TanStack Query v5, Keycloak JS |
| **Auth** | Keycloak, FreeMarker Templates (Theme) |
| **Infra** | Terraform, Kubernetes, Kustomize |
| **Database** | PostgreSQL |

---

## Directory Structure

```text
.
├── server/             # Go backend application
│   ├── cmd/            # Entry points
│   ├── internal/       # Business logic (Clean Architecture)
│   └── proto/          # API definitions (Protobuf)
├── ui/
│   ├── web/            # Vite + React frontend (Keycloak integration)
│   └── keycloak-theme/ # Keycloak custom theme
├── infra/
│   ├── tf/             # Terraform (IaC)
│   └── k8s/            # Kubernetes manifests
├── go.mod              # Go module definition
└── Makefile            # Project-wide task execution
```

Each directory has a detailed development guide (`AGENTS.md`).

---

## Getting Started

### Deployment to Kubernetes (MiniKube)

Manifests for MiniKube are available in `infra/k8s`.

```bash
# Generate manifests
kubectl kustomize infra/k8s/overlays/minikube

# Deploy
kubectl apply -k infra/k8s/overlays/minikube
```

Note: The `hub` image needs to be built beforehand.
Build it using the Docker daemon within MiniKube using `minikube docker-env`, or load the image into MiniKube.

### 1. Install Dependencies

```bash
# Backend development tools
make init

# Frontend dependent packages
cd ui/web && pnpm install
```

### 2. Start Local Development Environment

```bash
# Set up environment using Docker Compose and Terraform
make dev
```

### 3. Code Generation (Protobuf / SQL)

```bash
make gen
```

---

## Development Guidelines

Refer to the `AGENTS.md` in each directory for detailed guidelines of each component.

- [Backend Development Guide](server/AGENTS.md)
- [Frontend Development Guide](ui/web/AGENTS.md)
- [Keycloak Theme Development Guide](ui/keycloak-theme/AGENTS.md)
- [Infrastructure Development Guide (Terraform)](infra/tf/AGENTS.md)
- [Infrastructure Development Guide (Kubernetes)](infra/k8s/AGENTS.md)
