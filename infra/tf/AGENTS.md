# Terraform Tech Stack

- **Tooling**: Terraform (IaC)
- **Providers**: Keycloak (there is no AWS/GCP/cloud provider in this repo)
- **State Management**: Local backend (`backend "local" {}` in each environment's `backend.tf`)

### State Management Constraints

`dev` and `minikube` both use the Terraform `local` backend, which only
locks against concurrent processes on the same machine/filesystem — it
does **not** protect against two different machines (e.g. a developer's
laptop and the in-cluster `terraform-apply` Job defined in
`infra/k8s/overlays/<env>/terraform-job.yaml`) applying against the same
environment at the same time. Each has its own separate local state
(the Job's copy lives on the `terraform-state` PVC).

This project has no cloud provider, so a remote backend (S3/GCS/Terraform
Cloud) would mean provisioning cloud infrastructure solely to host
Terraform state — not justified for these single-operator, disposable
dev/minikube environments. Given that constraint:

- Treat `dev`/`minikube` as **single-operator environments**: do not run
  `terraform apply` from your machine while the in-cluster
  `terraform-apply` Job may also be running (or vice versa).
- If concurrent/multi-operator use becomes a real need, prefer the
  Terraform `kubernetes` backend (state stored as a Secret in the
  cluster, locked via a Lease) over a cloud backend — every environment
  here already requires `kubectl` cluster access, so it adds no new
  dependency.

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
