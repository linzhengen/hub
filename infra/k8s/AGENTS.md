# Kubernetes Tech Stack

- **Tooling**: Kustomize (native to `kubectl`)
- **Structure**: Base and Overlays (layered manifests)
- **Deployment Strategy**: GitOps ready (can be applied via `kubectl apply -k`)

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
