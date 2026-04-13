resource "keycloak_openid_client" "hub_api" {
  realm_id    = keycloak_realm.hub_realm.id
  client_id   = "hub-api"
  name        = "HUB API"
  enabled     = true
  access_type = "CONFIDENTIAL"
}
