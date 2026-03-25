resource "keycloak_realm" "hub_realm" {
  realm        = "hub"
  enabled      = true
  ssl_required = "none"

  login_theme = "hub"
}
