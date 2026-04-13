provider "keycloak" {
  username  = var.keycloak_user
  password  = var.keycloak_password
  url       = var.keycloak_url
  client_id = "admin-cli"
}
