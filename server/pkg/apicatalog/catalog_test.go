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

// GetMe is the one rpc that must stay reachable without a permission: a user
// with no roles yet still has to be able to load their own profile.
func TestGetMeIsTheOnlyPublicOperation(t *testing.T) {
	var public []string
	for _, op := range apicatalog.Default().Operations() {
		if op.Public {
			public = append(public, op.FullMethod)
		}
	}
	assert.Equal(t, []string{"/user.v1.UserService/GetMe"}, public)
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
	assert.Equal(t, []string{"group_id"}, add.PathParams())
	body, ok := add.Field("user_ids")
	require.True(t, ok)
	assert.Equal(t, apicatalog.InBody, body.In, `body: "*" puts every non-path field in the body`)
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
