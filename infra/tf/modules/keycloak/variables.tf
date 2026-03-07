variable "provider_github_client_id" {
  description = "Github app client id"
  sensitive   = true
}

variable "provider_github_client_secret" {
  description = "Github app client secret"
  sensitive   = true
}

variable "admin_user_name" {
  description = "Admin user name"
  default     = "admin"
}

variable "admin_user_email" {
  description = "Admin user email"
  default     = "admin@example.com"
}

variable "admin_user_password" {
  description = "Admin user password"
  type        = string
  sensitive   = true
}
