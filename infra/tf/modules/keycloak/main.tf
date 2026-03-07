resource "keycloak_realm" "hub_realm" {
  realm   = "hub"
  enabled = true
}
