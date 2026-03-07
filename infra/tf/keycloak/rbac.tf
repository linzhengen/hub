resource "keycloak_role" "admin_role" {
  realm_id = keycloak_realm.hub_realm.id
  name     = "admin"
}

resource "keycloak_user" "admin_user" {
  realm_id   = keycloak_realm.hub_realm.id
  username   = var.admin_user_name
  email      = var.admin_user_email
  first_name = "Admin"
  last_name  = "Istrator"

  dynamic "initial_password" {
    for_each = var.admin_user_password != "" ? [1] : []
    content {
      value     = var.admin_user_password
      temporary = false
    }
  }
}

resource "keycloak_user_roles" "admin_user_roles" {
  realm_id = keycloak_realm.hub_realm.id
  user_id  = keycloak_user.admin_user.id

  role_ids = [
    keycloak_role.admin_role.id
  ]
}
