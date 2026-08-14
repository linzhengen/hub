package apicatalog_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// The catalog is built from protoregistry.GlobalFiles, which only knows the
// generated packages something imported. That list used to be checked against a
// hand-written one here, so forgetting the blank import and forgetting the line
// in this test were the same mistake - which is how ai.chat.v1.ChatService came
// to be absent from the RBAC rules, the CLI, the web client's operation table
// and the agent reference at once, with every test still green.
//
// Reading proto/ instead means the expectation cannot be forgotten: a service
// declared in a .proto but not imported by catalog.go fails here.
func TestDefaultCoversEveryService(t *testing.T) {
	declared, err := servicesDeclaredInProtos(filepath.Join("..", "..", "proto"))
	require.NoError(t, err)
	require.NotEmpty(t, declared, "no .proto declared a service; the path is probably wrong")

	assert.ElementsMatch(t, declared, apicatalog.Default().Services(),
		"every service declared in proto/ needs a blank import in catalog.go")
}

var (
	protoPackage = regexp.MustCompile(`(?m)^package\s+([\w.]+)\s*;`)
	protoService = regexp.MustCompile(`(?m)^service\s+(\w+)\s*\{`)
)

// servicesDeclaredInProtos returns the fully qualified name of every service
// declared under root. It reads the .proto files rather than the descriptors so
// that it cannot share the omission it is checking for.
func servicesDeclaredInProtos(root string) ([]string, error) {
	var services []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".proto" {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		names := protoService.FindAllStringSubmatch(string(source), -1)
		if len(names) == 0 {
			return nil
		}
		pkg := protoPackage.FindStringSubmatch(string(source))
		if pkg == nil {
			return fmt.Errorf("%s declares a service but no package", path)
		}
		for _, name := range names {
			services = append(services, pkg[1]+"."+name[1])
		}
		return nil
	})
	return services, err
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

// GetMe, GetMeMenus, and SendMeVerifyEmail must stay reachable without a role
// assignment: a user with no roles yet still has to be able to load their own
// profile, navigation menu, and trigger email verification.  GetMeMenus returns
// an empty list when no roles are assigned, so making it public never leaks
// data — it just avoids a 403 that would leave the sidebar blank for brand-new
// users.  SendMeVerifyEmail is safe to make public because it always sends to
// the authenticated user's own address; no target id is accepted.
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
		"/user.v1.UserService/SendMeVerifyEmail",
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
