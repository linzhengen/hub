resource "keycloak_oidc_identity_provider" "github" {
  realm                 = keycloak_realm.hub_realm.id
  alias                 = "github"
  provider_id           = "github"
  client_id             = var.provider_github_client_id
  client_secret         = var.provider_github_client_secret
  enabled               = true
  backchannel_supported = false
  gui_order             = "1"
  store_token           = false
  sync_mode             = "LEGACY"
  default_scopes        = ""
  authorization_url     = ""
  token_url             = ""
}
