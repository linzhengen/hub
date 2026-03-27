resource "keycloak_realm" "hub_realm" {
  realm        = "hub"
  enabled      = true
  ssl_required = "none"

  login_theme = "hub"
  account_theme = "keycloak.v3"
  admin_theme = "keycloak.v2"
  email_theme = "keycloak"
}
