resource "keycloak_openid_client" "hub_web" {
  realm_id                            = keycloak_realm.hub_realm.id
  client_id                           = "hub-web"
  name                                = "HUB"
  enabled                             = true
  standard_flow_enabled               = true
  implicit_flow_enabled               = false
  direct_access_grants_enabled        = true
  backchannel_logout_session_required = true
  access_type                         = "PUBLIC"
  access_token_lifespan               = "1800"
  valid_redirect_uris = [
    "http://localhost:3000/*"
  ]
  web_origins = [
    "http://localhost:3000"
  ]
}

locals {
  roles = {
    admin   = "Administrator role with full privileges."
    manager = "Manager role with restricted administrative privileges."
    guest   = "Guest role with minimal access."
    editor  = "Editor role with content creation and editing privileges."
    viewer  = "Read-only viewer role with no modification privileges."
  }
}

resource "keycloak_role" "hub_web_roles" {
  for_each    = local.roles
  realm_id    = keycloak_realm.hub_realm.id
  client_id   = keycloak_openid_client.hub_web.id
  name        = each.key
  description = each.value
}
