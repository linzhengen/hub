package apicatalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

func TestDefaultCoversEveryService(t *testing.T) {
	assert.ElementsMatch(t, []string{
		"system.group.v1.GroupService",
		"system.permission.v1.PermissionService",
		"system.resource.v1.ResourceService",
		"system.role.v1.RoleService",
		"user.v1.UserService",
	}, apicatalog.Default().Services())
}

func TestOperationCarriesRestMappingAndRbacRule(t *testing.T) {
	op, ok := apicatalog.Default().ByFullMethod("/user.v1.UserService/ListUser")
	require.True(t, ok)

	assert.Equal(t, "GET", op.HTTPMethod)
	assert.Equal(t, "/api/v1/users", op.Path)
	assert.Equal(t, "api.user.v1.UserService", op.Resource)
	assert.Equal(t, "ListUser", op.Action)
	assert.False(t, op.Public)
	assert.NotEmpty(t, op.Summary, "every rpc should describe itself for the CLI and agent docs")
}

// GetMe and GetMeMenus must stay reachable without a role assignment: a user
// with no roles yet still has to be able to load their own profile and
// navigation menu.  GetMeMenus returns an empty list when no roles are
// assigned, so making it public never leaks data — it just avoids a 403 that
// would leave the sidebar blank for brand-new users.
func TestPublicOperations(t *testing.T) {
	var public []string
	for _, op := range apicatalog.Default().Operations() {
		if op.Public {
			public = append(public, op.FullMethod)
		}
	}
	assert.ElementsMatch(t, []string{
		"/user.v1.UserService/GetMe",
		"/user.v1.UserService/GetMeMenus",
	}, public)
}

func TestFieldsAreSplitBetweenPathQueryAndBody(t *testing.T) {
	catalog := apicatalog.Default()

	get, ok := catalog.ByFullMethod("/user.v1.UserService/GetUser")
	require.True(t, ok)
	assert.Equal(t, []string{"id"}, get.PathParams())

	list, ok := catalog.ByFullMethod("/user.v1.UserService/ListUser")
	require.True(t, ok)
	userIds, ok := list.Field("userIds")
	require.True(t, ok, "fields are addressable by their JSON name")
	assert.Equal(t, "user_ids", userIds.Name)
	assert.True(t, userIds.Repeated)
	assert.Equal(t, apicatalog.InQuery, userIds.In, "a GET has no body, so filters travel in the query")

	status, ok := list.Field("status")
	require.True(t, ok)
	assert.Equal(t, []string{"STATUS_UNSPECIFIED", "STATUS_ACTIVE", "STATUS_INACTIVE"}, status.EnumValues)

	add, ok := catalog.ByFullMethod("/system.group.v1.GroupService/AddUsersToGroup")
	require.True(t, ok)
	assert.Equal(t, []string{"id"}, add.PathParams(), "every path parameter naming the parent is called id")
	body, ok := add.Field("user_ids")
	require.True(t, ok)
	assert.Equal(t, apicatalog.InBody, body.In, `body: "*" puts every non-path field in the body`)
}

// The rules are surfaced so a caller - an agent above all - can get a request
// right before sending it, rather than learning by rejection.
func TestFieldsCarryTheirValidationRules(t *testing.T) {
	catalog := apicatalog.Default()

	create, ok := catalog.ByFullMethod("/user.v1.UserService/CreateUser")
	require.True(t, ok)

	for name, want := range map[string][]string{
		"username":  {"length 1..64"},
		"email":     {"email"},
		"password":  {"length 8..128"},
		"group_ids": {"each uuid"},
	} {
		field, ok := create.Field(name)
		require.True(t, ok, name)
		assert.Equal(t, want, field.Constraints, name)
	}

	list, ok := catalog.ByFullMethod("/user.v1.UserService/ListUser")
	require.True(t, ok)
	limit, ok := list.Field("limit")
	require.True(t, ok)
	assert.Equal(t, []string{"<= 200"}, limit.Constraints)

	offset, ok := list.Field("offset")
	require.True(t, ok)
	assert.Empty(t, offset.Constraints, "an unconstrained field reports nothing")

	add, ok := catalog.ByFullMethod("/system.group.v1.GroupService/AddRolesToGroup")
	require.True(t, ok)
	roleIDs, ok := add.Field("role_ids")
	require.True(t, ok)
	assert.Equal(t, []string{"at least 1 item(s)", "each uuid"}, roleIDs.Constraints)
}

// Every id a caller passes should be checked, or a typo becomes a confusing
// "not found" instead of a clear rejection.
func TestEveryIdPathParameterIsConstrained(t *testing.T) {
	for _, op := range apicatalog.Default().Operations() {
		for _, field := range op.Fields {
			if field.In != apicatalog.InPath {
				continue
			}
			assert.Contains(t, field.Constraints, "uuid",
				"%s.%s takes %s in the path with no uuid rule", op.Service, op.Method, field.Name)
		}
	}
}

func TestFind(t *testing.T) {
	catalog := apicatalog.Default()

	for _, ref := range []string{
		"/user.v1.UserService/ListUser",
		"user.v1.UserService.ListUser",
		"UserService.ListUser",
		"ListUser",
	} {
		op, err := catalog.Find(ref)
		require.NoError(t, err, ref)
		assert.Equal(t, "/user.v1.UserService/ListUser", op.FullMethod, ref)
	}

	_, err := catalog.Find("NoSuchMethod")
	assert.ErrorContains(t, err, "unknown operation")
}
