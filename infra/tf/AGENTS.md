# Terraform Tech Stack

- **Tooling**: Terraform (IaC)
- **Providers**: AWS/Google Cloud (assumed based on TF), Kubernetes, Keycloak
- **State Management**: Local or Remote Backend (S3/GCS as configured in `backend.tf`)

## Directory Structure

- `modules/`: Reusable Terraform modules.
  - `keycloak/`: Custom module for Keycloak configurations.
- `keycloak/`: Resources related to Keycloak realm/client management.
- `dev/`: Development environment setup.
  - `main.tf`: Entry point for dev resources.
  - `variables.tf`: Configuration variables.
  - `terraform.tfvars`: Environment-specific values.

## Development Workflow

### Infrastructure Provisioning
1. Initialize Terraform: `terraform init` (run inside an environment folder like `dev/`).
2. Plan changes: `terraform plan`.
3. Apply changes: `terraform apply`.

### Module Usage
- Create or update modules in `modules/`.
- Call modules from environment-specific configurations (`dev/main.tf`).
- Use variables and outputs to pass data between resources.

## Testing & Quality
- Always run `terraform fmt` for consistent formatting.
- Perform `terraform plan` to verify intended changes before applying.
- Use workspaces or separate directories for different environments.
- Follow the project-wide TDD principles (plan/apply cycle).
