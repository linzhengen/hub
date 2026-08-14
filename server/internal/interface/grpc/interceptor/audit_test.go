package interceptor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc"

	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// TestEveryMutationIsClassified is the guarantee that nothing changes the
// authorization graph unrecorded by accident.
//
// A new rpc reaches the catalog the moment its proto is generated, so an rpc
// added without a decision about auditing fails here rather than being found
// missing from the log during an investigation.
func TestEveryMutationIsClassified(t *testing.T) {
	catalog := apicatalog.Default()
	if missing := UnclassifiedMutations(catalog); len(missing) > 0 {
		t.Errorf("these rpcs change state but are in neither audited nor notAudited:\n  %s\n"+
			"add each to audited (it changes the authorization graph) or to notAudited (with the reason).",
			strings.Join(missing, "\n  "))
	}
}

// TestAuditedRpcsExistAndMutate catches the opposite drift: an entry left
// behind after the rpc was renamed or turned into a read.
func TestAuditedRpcsExistAndMutate(t *testing.T) {
	catalog := apicatalog.Default()
	for _, classification := range []map[string]bool{audited, notAudited} {
		for full := range classification {
			op, ok := catalog.ByFullMethod(full)
			if !ok {
				t.Errorf("%s is classified but no longer exists in the API catalog", full)
				continue
			}
			if !mutates(op) {
				t.Errorf("%s is classified as a mutation but is a %s", full, op.HTTPMethod)
			}
		}
	}
}

type recordingRepo struct {
	entries []*audit.Entry
	err     error
}

func (r *recordingRepo) Create(_ context.Context, e *audit.Entry) error {
	if r.err != nil {
		return r.err
	}
	copied := *e
	r.entries = append(r.entries, &copied)
	return nil
}

// passThroughTrans runs the callback without a database, which is what lets the
// interceptor be tested on its own. It reports whether it was used.
type passThroughTrans struct{ used bool }

func (t *passThroughTrans) ExecTrans(ctx context.Context, fn func(context.Context) error) error {
	t.used = true
	return fn(ctx)
}

func (t *passThroughTrans) ExecTransWithLock(ctx context.Context, fn func(context.Context) error) error {
	return t.ExecTrans(ctx, fn)
}

const (
	createUserMethod = "/user.v1.UserService/CreateUser"
	addRolesMethod   = "/system.group.v1.GroupService/AddRolesToGroup"
	listUserMethod   = "/user.v1.UserService/ListUser"
)

func newAuditFixture(t *testing.T) (*recordingRepo, *passThroughTrans, grpc.UnaryServerInterceptor) {
	t.Helper()
	catalog := apicatalog.Default()
	repo := &recordingRepo{}
	trans := &passThroughTrans{}
	return repo, trans, UnaryAuditInterceptor(repo, trans, catalog)
}

func callerContext() context.Context {
	return contextx.WithUserID(context.Background(), "11111111-1111-1111-1111-111111111111")
}

func TestAuditRecordsASuccessfulChange(t *testing.T) {
	repo, trans, intercept := newAuditFixture(t)

	req := &pbgroupv1.AddRolesToGroupRequest{
		Id:      "22222222-2222-2222-2222-222222222222",
		RoleIds: []string{"33333333-3333-3333-3333-333333333333"},
	}
	_, err := intercept(callerContext(), req,
		&grpc.UnaryServerInfo{FullMethod: addRolesMethod},
		func(ctx context.Context, _ any) (any, error) {
			return &pbgroupv1.AddRolesToGroupResponse{}, nil
		})
	if err != nil {
		t.Fatalf("interceptor returned %v", err)
	}

	if len(repo.entries) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repo.entries))
	}
	entry := repo.entries[0]
	if !entry.Succeeded {
		t.Error("entry is not marked as succeeded")
	}
	if entry.ActorUserId != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("actor is %q", entry.ActorUserId)
	}
	if entry.Channel != audit.ChannelAPI {
		t.Errorf("channel is %q, want %q", entry.Channel, audit.ChannelAPI)
	}
	if entry.TargetId != req.Id {
		t.Errorf("target is %q, want the group id %q", entry.TargetId, req.Id)
	}
	if entry.Action != "AddRolesToGroup" {
		t.Errorf("action is %q", entry.Action)
	}
	if !strings.Contains(string(entry.Arguments), "33333333-3333-3333-3333-333333333333") {
		t.Errorf("arguments do not mention the role that was granted: %s", entry.Arguments)
	}
	if !trans.used {
		t.Error("the record was not written in the handler's transaction")
	}
}

// A refused call is the event the log exists for, so it has to be recorded even
// though nothing changed.
func TestAuditRecordsARefusedAttempt(t *testing.T) {
	repo, _, intercept := newAuditFixture(t)

	refused := errors.New("permission denied for AddRolesToGroup")
	_, err := intercept(callerContext(), &pbgroupv1.AddRolesToGroupRequest{Id: "22222222-2222-2222-2222-222222222222"},
		&grpc.UnaryServerInfo{FullMethod: addRolesMethod},
		func(context.Context, any) (any, error) { return nil, refused })
	if !errors.Is(err, refused) {
		t.Fatalf("interceptor returned %v, want the handler's error", err)
	}

	if len(repo.entries) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repo.entries))
	}
	if repo.entries[0].Succeeded {
		t.Error("a refused attempt is recorded as succeeded")
	}
	if !strings.Contains(repo.entries[0].Error, "permission denied") {
		t.Errorf("the refusal is not recorded: %q", repo.entries[0].Error)
	}
}

// A create names its target only in the response.
func TestAuditTakesTheTargetFromACreateResponse(t *testing.T) {
	repo, _, intercept := newAuditFixture(t)

	created := "44444444-4444-4444-4444-444444444444"
	_, err := intercept(callerContext(), &pbuserv1.CreateUserRequest{Username: "someone"},
		&grpc.UnaryServerInfo{FullMethod: createUserMethod},
		func(context.Context, any) (any, error) {
			return &pbuserv1.CreateUserResponse{User: &pbuserv1.User{Id: created}}, nil
		})
	if err != nil {
		t.Fatalf("interceptor returned %v", err)
	}
	if len(repo.entries) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repo.entries))
	}
	if repo.entries[0].TargetId != created {
		t.Errorf("target is %q, want the created user %q", repo.entries[0].TargetId, created)
	}
}

func TestAuditNeverStoresAPassword(t *testing.T) {
	repo, _, intercept := newAuditFixture(t)

	_, err := intercept(callerContext(),
		&pbuserv1.CreateUserRequest{Username: "someone", Email: "someone@example.com", Password: "hunter2"},
		&grpc.UnaryServerInfo{FullMethod: createUserMethod},
		func(context.Context, any) (any, error) {
			return &pbuserv1.CreateUserResponse{User: &pbuserv1.User{Id: "44444444-4444-4444-4444-444444444444"}}, nil
		})
	if err != nil {
		t.Fatalf("interceptor returned %v", err)
	}
	if len(repo.entries) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repo.entries))
	}

	arguments := string(repo.entries[0].Arguments)
	if strings.Contains(arguments, "hunter2") {
		t.Errorf("the password was written to the audit log: %s", arguments)
	}
	// The rest of the request still has to be there, or the record explains
	// nothing.
	if !strings.Contains(arguments, "someone") {
		t.Errorf("the arguments were dropped entirely: %s", arguments)
	}
	var decoded map[string]any
	if err := json.Unmarshal(repo.entries[0].Arguments, &decoded); err != nil {
		t.Fatalf("arguments are not valid JSON: %v", err)
	}
	if _, present := decoded["password"]; present {
		t.Error("the password field is present, even if emptied")
	}
}

// The assistant is a channel, not an actor: the record still names the person.
func TestAuditRecordsTheChannelTheAssistantActedThrough(t *testing.T) {
	repo, _, intercept := newAuditFixture(t)

	ctx := audit.NewContext(callerContext(), audit.Source{
		Channel:   audit.ChannelAIChat,
		SessionId: "55555555-5555-5555-5555-555555555555",
	})
	_, err := intercept(ctx, &pbgroupv1.AddRolesToGroupRequest{Id: "22222222-2222-2222-2222-222222222222"},
		&grpc.UnaryServerInfo{FullMethod: addRolesMethod},
		func(context.Context, any) (any, error) { return &pbgroupv1.AddRolesToGroupResponse{}, nil })
	if err != nil {
		t.Fatalf("interceptor returned %v", err)
	}
	if len(repo.entries) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.Channel != audit.ChannelAIChat {
		t.Errorf("channel is %q, want %q", entry.Channel, audit.ChannelAIChat)
	}
	if entry.SessionId != "55555555-5555-5555-5555-555555555555" {
		t.Errorf("session is %q", entry.SessionId)
	}
	if entry.ActorUserId != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("actor is %q, want the person the assistant answered", entry.ActorUserId)
	}
}

func TestAuditIgnoresReads(t *testing.T) {
	repo, trans, intercept := newAuditFixture(t)

	_, err := intercept(callerContext(), &pbuserv1.ListUserRequest{},
		&grpc.UnaryServerInfo{FullMethod: listUserMethod},
		func(context.Context, any) (any, error) { return &pbuserv1.ListUserResponse{}, nil })
	if err != nil {
		t.Fatalf("interceptor returned %v", err)
	}
	if len(repo.entries) != 0 {
		t.Errorf("recorded %d entries for a read", len(repo.entries))
	}
	if trans.used {
		t.Error("a read was wrapped in a transaction")
	}
}

// The change and its record commit together, so a record that cannot be written
// takes the change down with it rather than leaving it unexplained.
func TestAuditFailureRollsTheChangeBack(t *testing.T) {
	catalog := apicatalog.Default()
	repo := &recordingRepo{err: errors.New("audit_logs is unavailable")}
	intercept := UnaryAuditInterceptor(repo, &passThroughTrans{}, catalog)

	_, err := intercept(callerContext(), &pbgroupv1.AddRolesToGroupRequest{Id: "22222222-2222-2222-2222-222222222222"},
		&grpc.UnaryServerInfo{FullMethod: addRolesMethod},
		func(context.Context, any) (any, error) { return &pbgroupv1.AddRolesToGroupResponse{}, nil })
	if err == nil {
		t.Fatal("the call succeeded even though its record could not be written")
	}
	if !strings.Contains(err.Error(), "audit_logs is unavailable") {
		t.Errorf("error is %v, want the audit failure", err)
	}
}
