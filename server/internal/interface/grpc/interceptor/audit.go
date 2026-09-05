package interceptor

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
	"github.com/linzhengen/hub/server/pkg/logger"
)

// audited lists the rpcs whose every attempt is written to the audit log: the
// ones that change the authorization graph.
//
// It is an explicit list rather than a rule such as "anything that is not a
// GET", because whether an operation has to be explainable afterwards is a
// judgement, not a property of its HTTP verb. TestEveryMutationIsClassified
// fails when a new mutating rpc appears in neither this map nor notAudited, so
// the judgement is made when the rpc is added rather than discovered missing
// during an investigation.
var audited = map[string]bool{
	"/system.access.v1.AccessRequestService/DecideAccessRequest": true,
	// Registering a machine, rotating its secret and removing it are each a
	// change to who can call the API, so each is recorded like any other.
	"/system.serviceaccount.v1.ServiceAccountService/CreateServiceAccount":       true,
	"/system.serviceaccount.v1.ServiceAccountService/RotateServiceAccountSecret": true,
	"/system.serviceaccount.v1.ServiceAccountService/DeleteServiceAccount":       true,
	// An agent is a machine identity too. Registering one, rotating its
	// credential and removing it are each a change to who can call the API.
	//
	// This is also the only place a *person* appears in an agent's story. An
	// agent acts as itself, so its own calls are recorded against the agent and
	// nothing says who asked it; the record of who registered it is therefore
	// what an investigation has to start from, until delegation gives the audit
	// log a chain to write.
	"/ai.agent.v1.AgentService/CreateAgent":       true,
	"/ai.agent.v1.AgentService/RotateAgentSecret": true,
	"/ai.agent.v1.AgentService/DeleteAgent":       true,
	// A delegation is the one thing that lets a machine borrow a person's
	// authority, so both ends of its life are recorded: granting it is the
	// event an investigation works back to, and revoking it is what somebody
	// will need to prove happened when they say they stopped it.
	"/system.delegation.v1.DelegationService/CreateDelegation": true,
	"/system.delegation.v1.DelegationService/RevokeDelegation": true,
	"/user.v1.UserService/CreateUser":                          true,
	"/user.v1.UserService/UpdateUser":                          true,
	"/user.v1.UserService/DeleteUser":                          true,
	"/user.v1.UserService/UpdateMe":                            true,
	"/user.v1.UserService/AddGroupsToUser":                     true,
	"/user.v1.UserService/RemoveGroupsFromUser":                true,
	// An organization is the boundary a permission is held within, so creating,
	// renaming, deactivating or deleting one changes what every grant inside it
	// reaches. Deleting one takes every group in it with it, which is the
	// largest single revocation the API can perform.
	"/system.organization.v1.OrganizationService/CreateOrganization": true,
	"/system.organization.v1.OrganizationService/UpdateOrganization": true,
	"/system.organization.v1.OrganizationService/DeleteOrganization": true,
	"/system.group.v1.GroupService/CreateGroup":                      true,
	"/system.group.v1.GroupService/UpdateGroup":                      true,
	"/system.group.v1.GroupService/DeleteGroup":                      true,
	"/system.group.v1.GroupService/AddRolesToGroup":                  true,
	"/system.group.v1.GroupService/RemoveRolesFromGroup":             true,
	"/system.group.v1.GroupService/AddUsersToGroup":                  true,
	"/system.group.v1.GroupService/RemoveUsersFromGroup":             true,
	"/system.role.v1.RoleService/CreateRole":                         true,
	"/system.role.v1.RoleService/UpdateRole":                         true,
	"/system.role.v1.RoleService/DeleteRole":                         true,
	"/system.role.v1.RoleService/AddPermissionsToRole":               true,
	"/system.role.v1.RoleService/RemovePermissionsFromRole":          true,
	"/system.permission.v1.PermissionService/CreatePermission":       true,
	"/system.permission.v1.PermissionService/UpdatePermission":       true,
	"/system.permission.v1.PermissionService/DeletePermission":       true,
	"/system.resource.v1.ResourceService/CreateResource":             true,
	"/system.resource.v1.ResourceService/UpdateResource":             true,
	"/system.resource.v1.ResourceService/DeleteResource":             true,
	"/system.resource.v1.ResourceService/CreateMenuResource":         true,
	"/system.resource.v1.ResourceService/UpdateMenuResource":         true,
}

// notAudited lists the mutating rpcs deliberately left out, so that leaving one
// out is a decision on the record rather than an omission.
//
// The chat rpcs are excluded because they change no permission and their
// arguments are the user's own prompt: recording them would build a transcript
// of everything anyone typed into the assistant, in a table read during
// investigations. What the assistant *does* on the user's behalf is audited -
// those calls arrive here as the rpc they actually invoke, marked with the
// ai_chat channel.
// ConfirmToolCall is excluded for a different reason: it is already accounted
// for at both ends. An approval that let a change through is recorded on that
// change, as its approval_id; a decline is recorded on the proposal row, which
// keeps its status and the time it was decided. A record here would duplicate
// the first and add nothing to the second.
// Raising and withdrawing a request are excluded because the request row is
// itself the record: it keeps who asked, for whom, why, when, and what became
// of it, and unlike a log entry it is what the decision is made from. A second
// copy in audit_logs would say less and drift. The decision is audited, because
// that is the point at which the graph changes.
var notAudited = map[string]bool{
	"/system.access.v1.AccessRequestService/CreateAccessRequest": true,
	"/system.access.v1.AccessRequestService/CancelAccessRequest": true,
	"/ai.chat.v1.ChatService/CreateSession":                      true,
	"/ai.chat.v1.ChatService/DeleteSession":                      true,
	"/ai.chat.v1.ChatService/SendMessage":                        true,
	"/ai.chat.v1.ChatService/ConfirmToolCall":                    true,
	"/user.v1.UserService/SendMeVerifyEmail":                     true,
}

// UnclassifiedMutations returns the mutating rpcs that appear in neither
// audited nor notAudited, newest additions to the API being the usual cause.
//
// An rpc that changes something and is in neither map is not a silent decision
// to skip it - it is nobody having decided. TestEveryMutationIsClassified fails
// on a non-empty result, and the server reports one at startup, so the gap is
// visible from CI and from a running deployment rather than only from an
// investigation that later finds nothing recorded.
func UnclassifiedMutations(catalog *apicatalog.Catalog) []string {
	if catalog == nil {
		return nil
	}
	var missing []string
	for _, op := range catalog.Operations() {
		if !mutates(op) {
			continue
		}
		if audited[op.FullMethod] || notAudited[op.FullMethod] {
			continue
		}
		missing = append(missing, op.FullMethod)
	}
	sort.Strings(missing)
	return missing
}

// mutates reports whether an rpc changes state, by its REST verb. An rpc with
// no REST mapping is treated as mutating so that adding one cannot slip past
// the classification by having no verb to read.
func mutates(op apicatalog.Operation) bool {
	return op.HTTPMethod != http.MethodGet
}

// secret lists request fields that must never reach the log, at any depth. A
// password is the only thing here today; the point of the list is that it is
// the one place to add the next one.
var secret = map[string]bool{
	"password":     true,
	"newPassword":  true,
	"new_password": true,
}

// UnaryAuditInterceptor records every attempt at an audited rpc.
//
// It runs before authorization, so a refused attempt is recorded too: someone
// trying to grant themselves a role and being told no is precisely the event
// the log exists to show.
//
// A successful call and its record share one transaction. The record is written
// inside the same transaction the use cases below join, so the change and the
// account of it commit together - a change cannot be left in the database with
// no record of who made it. A failed attempt is recorded separately, since the
// transaction carrying it has been rolled back.
func UnaryAuditInterceptor(
	auditRepo audit.Repository,
	transRepo trans.Repository,
	catalog *apicatalog.Catalog,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !audited[info.FullMethod] {
			return handler(ctx, req)
		}
		op, ok := catalog.ByFullMethod(info.FullMethod)
		if !ok {
			// An audited rpc the catalog does not know cannot be described in
			// the RBAC vocabulary the log is written in. Failing the call would
			// turn a metadata gap into an outage, so it proceeds and says so.
			logger.Errorf("audit: %s is audited but not in the API catalog", info.FullMethod)
			return handler(ctx, req)
		}
		actor, ok := contextx.GetUserID(ctx)
		if !ok {
			// Unauthenticated: there is no actor to hold responsible, and
			// authorization is about to reject the call anyway.
			return handler(ctx, req)
		}

		entry := newEntry(ctx, op, req, actor)

		var resp any
		err := transRepo.ExecTrans(ctx, func(txCtx context.Context) error {
			var handlerErr error
			resp, handlerErr = handler(txCtx, req)
			if handlerErr != nil {
				return handlerErr
			}
			entry.Succeeded = true
			entry.TargetId = targetId(req, resp)
			return auditRepo.Create(txCtx, entry)
		})
		if err != nil {
			entry.Succeeded = false
			entry.Error = err.Error()
			recordAttempt(ctx, auditRepo, entry)
			return nil, err
		}
		return resp, nil
	}
}

// newEntry describes the attempt from what is known before it runs.
func newEntry(ctx context.Context, op apicatalog.Operation, req any, actor string) *audit.Entry {
	source := audit.FromContext(ctx)
	return &audit.Entry{
		ActorUserId: actor,
		// The organization the caller named, empty when they named none. Taken
		// from the context rather than from the request, because it qualifies
		// every rpc rather than being a field on any of them.
		OrgId:      contextx.GetOrgID(ctx),
		Channel:    source.Channel,
		SessionId:  source.SessionId,
		Resource:   op.Resource,
		Action:     op.Action,
		TargetId:   targetId(req, nil),
		Arguments:  arguments(req),
		Client:     clientOf(ctx),
		ApprovalId: source.ApprovalId,
	}
}

// recordAttempt writes a failed attempt.
//
// ctx here is the interceptor's own context, not the one ExecTrans derived, so
// the write does not join the transaction that was just rolled back. It cannot
// fail the call - the call has already failed - so a problem is reported to the
// log and the error returned to the caller is still the handler's.
func recordAttempt(ctx context.Context, auditRepo audit.Repository, entry *audit.Entry) {
	if err := auditRepo.Create(ctx, entry); err != nil {
		logger.Errorf("audit: failed to record %s %s by %s: %v",
			entry.Action, entry.Resource, entry.ActorUserId, err)
	}
}

// arguments renders the request as JSON with the secret fields removed.
func arguments(req any) []byte {
	message, ok := req.(proto.Message)
	if !ok {
		return nil
	}
	encoded, err := protojson.Marshal(message)
	if err != nil {
		logger.Errorf("audit: failed to marshal %T: %v", req, err)
		return nil
	}
	var body any
	if err := json.Unmarshal(encoded, &body); err != nil {
		logger.Errorf("audit: failed to decode %T: %v", req, err)
		return nil
	}
	scrubbed, err := json.Marshal(scrub(body))
	if err != nil {
		logger.Errorf("audit: failed to encode %T: %v", req, err)
		return nil
	}
	return scrubbed
}

// scrub walks the decoded request and drops the secret fields at any depth. It
// walks the JSON rather than the message so a secret field added to a nested
// message is covered without a per-message list to keep up to date.
func scrub(node any) any {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if secret[key] {
				delete(value, key)
				continue
			}
			value[key] = scrub(child)
		}
		return value
	case []any:
		for i, child := range value {
			value[i] = scrub(child)
		}
		return value
	default:
		return node
	}
}

// targetId is the id of the entity the call changed.
//
// An update, a delete or a link names it in the request; a create only learns
// it from the response, where it sits one message down - CreateGroupResponse
// carries the Group. Both are read reflectively so that a new rpc following the
// same convention needs no entry here.
func targetId(req, resp any) string {
	if id := idField(req); id != "" {
		return id
	}
	return nestedId(resp)
}

func idField(message any) string {
	m, ok := message.(proto.Message)
	if !ok {
		return ""
	}
	fields := m.ProtoReflect().Descriptor().Fields()
	field := fields.ByName("id")
	if field == nil || field.Kind() != protoreflect.StringKind || field.IsList() {
		return ""
	}
	return m.ProtoReflect().Get(field).String()
}

// nestedId reads the id of the single entity a response carries.
func nestedId(message any) string {
	m, ok := message.(proto.Message)
	if !ok {
		return ""
	}
	reflected := m.ProtoReflect()
	fields := reflected.Descriptor().Fields()
	for i := range fields.Len() {
		field := fields.Get(i)
		if field.Kind() != protoreflect.MessageKind || field.IsList() || field.IsMap() {
			continue
		}
		if id := idField(reflected.Get(field).Message().Interface()); id != "" {
			return id
		}
	}
	return ""
}

// clientOf reports what the caller said it was. It is a user agent, so it is
// not evidence of anything - it is recorded to tell the console apart from the
// CLI when reading the log, not to decide anything.
func clientOf(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if agents := md.Get("user-agent"); len(agents) > 0 {
		return agents[0]
	}
	return ""
}
