package tool_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/infrastructure/ai/tool"
	pbaccessv1 "github.com/linzhengen/hub/server/pb/system/access/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

const caller = "11111111-1111-1111-1111-111111111111"

// allowAll / denyAll stand in for the RBAC service. The point of the tests
// below is which tools are offered and whether a call gets through, not how a
// policy is matched - that is auth.Service's own suite.
type authStub struct {
	allow map[string]bool
	all   bool
}

func (s authStub) Enforce(_ context.Context, req auth.Request) (bool, error) {
	if s.all {
		return true, nil
	}
	return s.allow[req.Action], nil
}

// The tool box asks only whether a call is allowed, never why, so the explain
// half of the service is here to satisfy the interface.

func (s authStub) Explain(_ context.Context, _ auth.Request) ([]auth.AccessPath, error) {
	return nil, nil
}

func (s authStub) PrincipalsFor(_ context.Context, _ auth.Request) ([]auth.Principal, error) {
	return nil, nil
}

func (s authStub) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return auth.Scope{All: true}, nil
}

// userServer is a stand-in for the real UserServiceServer, returning one user
// with an email so the redaction can be observed.
type userServer struct {
	listed int
}

func (s *userServer) ListUser(_ context.Context, _ *pbuserv1.ListUserRequest) (*pbuserv1.ListUserResponse, error) {
	s.listed++
	return &pbuserv1.ListUserResponse{Users: []*pbuserv1.User{{
		Id:       "22222222-2222-2222-2222-222222222222",
		Username: "taro",
		Email:    "taro@example.com",
	}}}, nil
}

func (s *userServer) GetUser(_ context.Context, req *pbuserv1.GetUserRequest) (*pbuserv1.GetUserResponse, error) {
	return &pbuserv1.GetUserResponse{User: &pbuserv1.User{Id: req.Id, Username: "taro", Email: "taro@example.com"}}, nil
}

func (s *userServer) CreateUser(context.Context, *pbuserv1.CreateUserRequest) (*pbuserv1.CreateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) UpdateUser(context.Context, *pbuserv1.UpdateUserRequest) (*pbuserv1.UpdateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) DeleteUser(context.Context, *pbuserv1.DeleteUserRequest) (*pbuserv1.DeleteUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) GetMe(context.Context, *pbuserv1.GetMeRequest) (*pbuserv1.GetMeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) GetMeMenus(context.Context, *pbuserv1.GetMeMenusRequest) (*pbuserv1.GetMeMenusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) UpdateMe(context.Context, *pbuserv1.UpdateMeRequest) (*pbuserv1.UpdateMeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) SendMeVerifyEmail(context.Context, *pbuserv1.SendMeVerifyEmailRequest) (*pbuserv1.SendMeVerifyEmailResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) AddGroupsToUser(context.Context, *pbuserv1.AddGroupsToUserRequest) (*pbuserv1.AddGroupsToUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}
func (s *userServer) RemoveGroupsFromUser(context.Context, *pbuserv1.RemoveGroupsFromUserRequest) (*pbuserv1.RemoveGroupsFromUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}

func ctxWithUser() context.Context {
	return contextx.WithUserID(context.Background(), caller)
}

func build(t *testing.T, authz auth.Service, interceptors ...grpc.UnaryServerInterceptor) (*userServer, chatDomain.ToolBox) {
	t.Helper()
	server := &userServer{}
	box, err := tool.New(
		apicatalog.Default(),
		authz,
		[]tool.Service{tool.Register(&pbuserv1.UserService_ServiceDesc, server)},
		interceptors...,
	)
	require.NoError(t, err)
	return server, box
}

// Only the operations on the allow list become tools, and only the ones the
// caller is permitted to run.
func TestToolsAreFilteredByPermission(t *testing.T) {
	_, box := build(t, authStub{all: true})
	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	names := toolNames(tools)
	assert.Equal(t, []string{"create_user", "get_user", "list_user", "update_user"}, names,
		"only the operations on the allow list are exposed")

	_, box = build(t, authStub{allow: map[string]bool{"ListUser": true}})
	tools, err = box.Tools(ctxWithUser())
	require.NoError(t, err)
	assert.Equal(t, []string{"list_user"}, toolNames(tools),
		"an operation the caller cannot run is not offered")

	_, box = build(t, authStub{})
	tools, err = box.Tools(ctxWithUser())
	require.NoError(t, err)
	assert.Empty(t, tools)
}

// Writes must not leak in through the allow list, whatever the caller may do.
// A tool that changes something must say so, because that flag is the only
// thing standing between the model asking and the change happening.
func TestWritesAreMarkedAsWrites(t *testing.T) {
	_, box := build(t, authStub{all: true})
	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	for _, tool := range tools {
		changes := false
		for _, verb := range []string{"create_", "update_", "delete_", "add_", "remove_"} {
			if strings.HasPrefix(tool.Name, verb) {
				changes = true
			}
		}
		assert.Equal(t, changes, tool.Write,
			"%s: name says it changes data = %v, but Write = %v", tool.Name, changes, tool.Write)
	}
}

// The escalation chain is excluded, so the rpcs that compose into it must not
// reach the model even though the caller is permitted to run them.
func TestTheEscalationChainIsNeverOffered(t *testing.T) {
	_, box := build(t, authStub{all: true})
	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	for _, name := range toolNames(tools) {
		for _, excluded := range []string{"add_groups_to_user", "remove_groups_from_user", "delete_user"} {
			assert.NotEqual(t, excluded, name, "%s reached the model despite being excluded", name)
		}
	}
}

// Deleting is on the excluded list, so no tool may delete.
func TestNoToolDeletes(t *testing.T) {
	_, box := build(t, authStub{all: true})
	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	for _, name := range toolNames(tools) {
		assert.False(t, strings.HasPrefix(name, "delete_"), "%s deletes", name)
	}
}

// The listing is a convenience; the interceptor is the guarantee. A tool the
// model asks for anyway still has to pass it.
func TestCallGoesThroughTheInterceptor(t *testing.T) {
	refused := status.Error(codes.PermissionDenied, "permission denied")
	deny := func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
		return nil, refused
	}

	server, box := build(t, authStub{all: true}, deny)
	_, err := box.Call(ctxWithUser(), "list_user", nil)

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
	assert.Zero(t, server.listed, "the service must not run when the interceptor refuses")
}

func TestCallReachesTheServiceAndHidesEmail(t *testing.T) {
	server, box := build(t, authStub{all: true})

	result, err := box.Call(ctxWithUser(), "list_user", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, server.listed)

	assert.Contains(t, result, "taro", "the name the answer needs is kept")
	assert.NotContains(t, result, "taro@example.com", "the address the answer does not need is dropped")
	assert.NotContains(t, result, "email")

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &body))
	users, ok := body["users"].([]any)
	require.True(t, ok)
	require.Len(t, users, 1)
	assert.NotContains(t, users[0].(map[string]any), "email", "redaction reaches nested messages")
}

func TestCallPassesArgumentsThrough(t *testing.T) {
	_, box := build(t, authStub{all: true})

	result, err := box.Call(ctxWithUser(), "get_user", []byte(`{"id":"33333333-3333-3333-3333-333333333333"}`))
	require.NoError(t, err)
	assert.Contains(t, result, "33333333-3333-3333-3333-333333333333")
}

func TestUnknownToolIsRejected(t *testing.T) {
	_, box := build(t, authStub{all: true})
	_, err := box.Call(ctxWithUser(), "drop_database", nil)
	assert.ErrorContains(t, err, "no such tool")
}

func TestToolsNeedAUser(t *testing.T) {
	_, box := build(t, authStub{all: true})
	_, err := box.Tools(context.Background())
	assert.Error(t, err)
}

// The schema is what the model writes arguments against, so it has to name the
// fields the rpc actually takes and mark the ones it cannot omit.
func TestSchemaComesFromTheCatalog(t *testing.T) {
	_, box := build(t, authStub{all: true})
	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	byName := map[string]chatDomain.Tool{}
	for _, tl := range tools {
		byName[tl.Name] = tl
	}

	get := byName["get_user"]
	assert.Equal(t, []string{"id"}, get.InputSchema["required"])
	properties := get.InputSchema["properties"].(map[string]any)
	assert.Contains(t, properties, "id")

	list := byName["list_user"]
	listProperties := list.InputSchema["properties"].(map[string]any)
	assert.Contains(t, listProperties, "limit")
	assert.Equal(t, "integer", listProperties["limit"].(map[string]any)["type"])
	assert.Equal(t, "array", listProperties["userIds"].(map[string]any)["type"],
		"a repeated field is an array so the model does not send a comma-joined string")
	assert.NotEmpty(t, list.Description, "the model needs to know what the tool does")
}

func toolNames(tools []chatDomain.Tool) []string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return names
}

// accessRequestServer is a stand-in for the real AccessRequestServiceServer.
// Only the two rpcs these tests reach do anything.
type accessRequestServer struct {
	created *pbaccessv1.CreateAccessRequestRequest
}

func (s *accessRequestServer) CreateAccessRequest(
	_ context.Context,
	req *pbaccessv1.CreateAccessRequestRequest,
) (*pbaccessv1.CreateAccessRequestResponse, error) {
	s.created = req
	return &pbaccessv1.CreateAccessRequestResponse{
		AccessRequest: &pbaccessv1.AccessRequest{Id: "33333333-3333-3333-3333-333333333333"},
	}, nil
}

func (s *accessRequestServer) ListAccessRequests(
	context.Context,
	*pbaccessv1.ListAccessRequestsRequest,
) (*pbaccessv1.ListAccessRequestsResponse, error) {
	return &pbaccessv1.ListAccessRequestsResponse{}, nil
}

func (s *accessRequestServer) CancelAccessRequest(
	context.Context,
	*pbaccessv1.CancelAccessRequestRequest,
) (*pbaccessv1.CancelAccessRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}

func (s *accessRequestServer) DecideAccessRequest(
	context.Context,
	*pbaccessv1.DecideAccessRequestRequest,
) (*pbaccessv1.DecideAccessRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not used")
}

func buildRequests(t *testing.T, authz auth.Service) chatDomain.ToolBox {
	t.Helper()
	box, err := tool.New(
		apicatalog.Default(),
		authz,
		[]tool.Service{tool.Register(
			&pbaccessv1.AccessRequestService_ServiceDesc, &accessRequestServer{})},
	)
	require.NoError(t, err)
	return box
}

// TestRaisingAccessIsOfferedToSomebodyWithNoPermissions is the behaviour the
// whole request flow depends on.
//
// The people who most need to ask for access are the ones holding none, so
// CreateAccessRequest is public and the authorization interceptor skips the
// check for it. The tool list has to agree: filtering it with Enforce anyway
// would withhold from those users a tool the server would have run for them.
func TestRaisingAccessIsOfferedToSomebodyWithNoPermissions(t *testing.T) {
	box := buildRequests(t, authStub{}) // permits nothing

	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	assert.Equal(t, []string{"create_access_request"}, toolNames(tools),
		"asking for access must not itself require a permission")
	// Listing requests is not public, so it stays behind the permission and is
	// absent here - which is what shows the filter is still doing its job.
}

// Deciding must not appear even for somebody permitted to do everything: it is
// on the escalation list, and that list does not bend to permissions.
func TestDecidingIsNotOfferedToAnyone(t *testing.T) {
	box := buildRequests(t, authStub{all: true})

	tools, err := box.Tools(ctxWithUser())
	require.NoError(t, err)

	for _, name := range toolNames(tools) {
		assert.NotEqual(t, "decide_access_request", name,
			"the assistant must never be offered a way to approve a request")
	}
	assert.Contains(t, toolNames(tools), "create_access_request")
	assert.Contains(t, toolNames(tools), "list_access_requests")
}
