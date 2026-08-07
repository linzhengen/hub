# Kubernetes Tech Stack

- **Tooling**: Kustomize (native to `kubectl`)
- **Structure**: Base and Overlays (layered manifests)
- **Deployment Strategy**: GitOps ready (can be applied via `kubectl apply -k`)

## Environment Scope

`base/` is currently written for the **dev/minikube** overlays only — the two
overlays that exist in `overlays/`. Both are disposable, single-operator
clusters reached over a port-forward, so `base/remote-charts/` deliberately
trades hardening for convenience. Treat `base/` as dev-only until an overlay
for a shared or production-like environment exists.

Before reusing `base/` in any other environment, an overlay must patch at least
the following (each is also flagged in a comment at the relevant line):

| Setting | Where | Why it must change outside dev |
| --- | --- | --- |
| `KC_HTTP_ENABLED: "true"` | `base/remote-charts/keycloak-values.yaml` | Tokens and admin credentials would travel in plaintext. Terminate TLS in front of Keycloak and drop this. |
| `KC_HOSTNAME_STRICT: "false"` | `base/remote-charts/keycloak-values.yaml` | Disables hostname validation, allowing Host-header-driven redirect/issuer values. Set `KC_HOSTNAME` to the public URL and flip this to `"true"`. |
| `primary.persistence.enabled: false` | `base/remote-charts/postgres-values.yaml` | All data — including the `keycloak` database and its realm/user records — is lost on every Pod restart or reschedule. Enable persistence with an explicit `storageClass`/`size`. |
| `global.security.allowInsecureImages: true` | both `*-values.yaml` | Bypasses the Bitnami chart's image check. It is set because this project ships its own rebuilds of those images; keep it only as long as that stays true, and pin digests where possible. |
| `*-secret.env.sample` | `base/` | Sample secrets are committed as templates. Real environments must supply values from a secret store — see "Secret Management" below. |

The in-cluster Terraform Job carries a similar caveat; see "In-Cluster Terraform
Jobs" at the end of this document.

## Directory Structure

- `base/`: Common manifests shared across all environments.
  - `kustomization.yaml`: Entry point for common resources.
  - `*-secret.env`: Template environment variables for secrets.
- `overlays/`: Environment-specific configurations.
  - `dev/`: Development environment overlay.
    - `kustomization.yaml`: Customizes or patches base manifests for dev.
    - `*-secret.env`: Environment-specific secrets.

## Secret Management

This project uses Kustomize `secretGenerator` to manage sensitive information.

1.  **Base Secrets**: Define default (usually dummy) values in `base/*.secret.env`.
2.  **Overlay Secrets**: Override values for specific environments in `overlays/<env>/*.secret.env`.
3.  **Consumption**:
    - Hub: Loaded via `envFrom` in the Helm chart.
    - Keycloak: Referenced via `existingSecret` in `keycloak-values.yaml`.
    - PostgreSQL: Referenced via `existingSecret` in `postgres-values.yaml`.

**Warning**: Do not commit actual production secrets to the repository. Use these files as templates and manage real secrets via a secure CI/CD pipeline or external secret store (e.g., SealedSecrets, ExternalSecrets).

## Development Workflow

### Manifest Management
1. Add common resources to `base/`.
2. Define environment-specific changes (e.g., replicas, image tags, env vars) in `overlays/<env>/`.
3. Use `kubectl kustomize overlays/<env>` to preview the final YAML.

### Deployment
1. To apply changes to an environment:
   ```bash
   kubectl apply -k overlays/<env>
   ```

## Testing & Quality
- Validate YAML syntax and Kustomize integrity before committing.
- Ensure proper resource limits and health checks are defined in `base/`.
- Follow the project-wide consistent naming conventions for resources.

## In-Cluster Terraform Jobs (`overlays/dev`, `overlays/minikube`)

`terraform-job.yaml` in these overlays runs `terraform apply -auto-approve`
inside the cluster with no plan review, using credentials from a Secret
(`terraform-secret-dev` / `terraform-secret`). This is intentional for these
disposable, single-operator dev/minikube environments — see
`infra/tf/AGENTS.md`'s "State Management Constraints" section for why a
plan-review gate isn't used here and what the concurrent-apply risk is.
Do not copy this pattern to a shared or production-like environment without
adding a plan/approval step first.
