variable "keycloak_user" {
  description = "Keycloak client user for Terraform"
}

variable "keycloak_password" {
  description = "Keycloak client password for Terraform"
}

variable "keycloak_url" {
  description = "Keycloak URL"
}

variable "provider_github_client_id" {
  description = "Github app client id"
  sensitive   = true
}

variable "provider_github_client_secret" {
  description = "Github app client secret"
  sensitive   = true
}
