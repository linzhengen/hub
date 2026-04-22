module "keycloak" {
  source                        = "../modules/keycloak"
  provider_github_client_id     = var.provider_github_client_id
  provider_github_client_secret = var.provider_github_client_secret
  admin_user_password           = "admin"
}
