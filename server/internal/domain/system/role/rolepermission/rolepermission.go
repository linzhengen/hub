package rolepermission

type RolePermission struct {
	RoleId       string
	PermissionId string
}

type RolePermissions []*RolePermission
