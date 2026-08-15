// Package tool exposes hub's own read operations to the assistant.
//
// The operations, their arguments and their constraints all come from
// pkg/apicatalog, so a tool is described to the model with the same metadata
// the server routes and validates on rather than a second hand-written copy.
//
// A call is dispatched in-process through the generated gRPC handler, with the
// authorization and validation interceptors the server itself uses. That means
// the checks are not reimplemented here: there is one implementation of "may
// this user do this" and one of "is this request well formed", and the
// assistant goes through both. No HTTP request is made and no bearer token is
// held, which is why this is not routed back through the REST gateway.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// exposed lists the rpcs the assistant may call, by gRPC method path, and says
// whether each one writes.
//
// It is an explicit list rather than a rule such as "anything called Get* or
// List*", so that adding an rpc to the API never silently hands it to a model.
// A read runs when the model asks; a write is only ever proposed, and runs after
// the user has approved it.
var exposed = map[string]bool{
	"/user.v1.UserService/ListUser":                          false,
	"/user.v1.UserService/GetUser":                           false,
	"/system.group.v1.GroupService/ListGroup":                false,
	"/system.group.v1.GroupService/GetGroup":                 false,
	"/system.role.v1.RoleService/ListRole":                   false,
	"/system.role.v1.RoleService/GetRole":                    false,
	"/system.permission.v1.PermissionService/ListPermission": false,
	"/system.permission.v1.PermissionService/GetPermission":  false,
	"/system.resource.v1.ResourceService/ListResource":       false,
	"/system.resource.v1.ResourceService/GetResource":        false,
	// The audit log answers "who added them to that group, and when" without
	// anyone having to write SQL, which is the question it exists for.
	"/system.audit.v1.AuditService/ListAuditLog": false,
	// Explain turns "can they?" into "why can they?", which is the question
	// somebody actually asks when access looks wrong. Both read the same graph
	// the other reads, and neither can reach anything a list of groups and
	// roles could not.
	"/system.access.v1.AccessService/ExplainUserAccess":          false,
	"/system.access.v1.AccessService/ListPrincipalsForOperation": false,
	"/system.access.v1.AccessRequestService/ListAccessRequests":  false,

	// Writes. Each is proposed to the user and runs only once approved.
	"/user.v1.UserService/CreateUser":                    true,
	"/user.v1.UserService/UpdateUser":                    true,
	"/system.group.v1.GroupService/CreateGroup":          true,
	"/system.group.v1.GroupService/UpdateGroup":          true,
	"/system.group.v1.GroupService/AddUsersToGroup":      true,
	"/system.group.v1.GroupService/RemoveUsersFromGroup": true,
	// Raising a request is the assistant's way into the work the escalation
	// list keeps it out of. It grants nothing: it adds a pending row that a
	// different person has to approve, in a screen the conversation is not in.
	"/system.access.v1.AccessRequestService/CreateAccessRequest": true,
}

// escalation lists the rpcs the assistant may never call, whatever the user is
// permitted to do and whatever they approve.
//
// These three compose into a path to any permission at all:
//
//	AddPermissionsToRole -> AddRolesToGroup -> AddGroupsToUser
//
// Approval is not enough of a guard for them. The threat is not that the model
// exceeds the user's rights - it cannot - but that text written by somebody else
// and read back through a tool result talks the model into proposing a change,
// and the person approving it is the same person the injected text is written
// to persuade ("this is routine, approve it"). Changing who can do what is rare,
// deliberate work with the highest cost when it goes wrong and the least to gain
// from being automated, so it stays in the console where a human drives it.
//
// The deletes are here for the same reason: a deletion the user did not intend
// is not undone by having approved it.
//
// DecideAccessRequest is here for a sharper version of the same reason, and is
// the entry that makes the rest of this workable. The assistant may raise an
// access request; it may never decide one. That asymmetry is the whole design:
// injected text can get as far as a pending row with a reason on it, and no
// further, because the person who settles it does so somewhere the conversation
// cannot reach and after reading what was actually asked for. Approving through
// the tool flow would collapse the two back into one screen and one person, and
// the approval would attest to nothing.
var escalation = map[string]bool{
	"/system.access.v1.AccessRequestService/DecideAccessRequest": true,
	"/system.role.v1.RoleService/AddPermissionsToRole":           true,
	"/system.role.v1.RoleService/RemovePermissionsFromRole":      true,
	"/system.group.v1.GroupService/AddRolesToGroup":              true,
	"/system.group.v1.GroupService/RemoveRolesFromGroup":         true,
	"/user.v1.UserService/AddGroupsToUser":                       true,
	"/user.v1.UserService/RemoveGroupsFromUser":                  true,
	"/user.v1.UserService/DeleteUser":                            true,
	"/system.group.v1.GroupService/DeleteGroup":                  true,
	"/system.role.v1.RoleService/DeleteRole":                     true,
	"/system.permission.v1.PermissionService/DeletePermission":   true,
	"/system.resource.v1.ResourceService/DeleteResource":         true,
}

// hidden lists response fields the assistant never needs, removed from every
// result before the model sees it.
//
// email is the one that matters: answering "who is in the admin group" needs a
// name, not an address, and an address is the field most directly usable for
// contacting or impersonating someone. The timestamps are simply noise. Both
// json and protobuf spellings are listed because the field name a result
// carries depends on how it was marshalled.
var hidden = map[string]bool{
	"email":      true,
	"createdAt":  true,
	"created_at": true,
	"updatedAt":  true,
	"updated_at": true,
}

// Service pairs a generated service description with the implementation the DI
// container built, which is what lets a call be dispatched by method name.
type Service struct {
	desc *grpc.ServiceDesc
	impl any
}

// Register names a service the assistant may reach. desc is the generated
// ServiceDesc, impl the server the DI container built for it.
func Register(desc *grpc.ServiceDesc, impl any) Service {
	return Service{desc: desc, impl: impl}
}

type toolBox struct {
	catalog     *apicatalog.Catalog
	authSvc     auth.Service
	interceptor grpc.UnaryServerInterceptor
	// byName maps a tool name to the operation and service that answer it.
	byName map[string]binding
}

type binding struct {
	op      apicatalog.Operation
	svc     Service
	handler methodHandler
	write   bool
}

type methodHandler func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error)

// New builds the tool box from the registered services.
//
// interceptors are applied to every dispatched call in order; pass the server's
// own authorization and validation interceptors so a tool call is checked
// exactly as the equivalent API request would be.
func New(
	catalog *apicatalog.Catalog,
	authSvc auth.Service,
	services []Service,
	interceptors ...grpc.UnaryServerInterceptor,
) (chatDomain.ToolBox, error) {
	box := &toolBox{
		catalog:     catalog,
		authSvc:     authSvc,
		interceptor: grpcmiddleware.ChainUnaryServer(interceptors...),
		byName:      map[string]binding{},
	}

	for _, svc := range services {
		for i := range svc.desc.Methods {
			method := svc.desc.Methods[i]
			full := "/" + svc.desc.ServiceName + "/" + method.MethodName
			write, listed := exposed[full]
			if !listed {
				continue
			}
			// Refusing here rather than trusting the two maps not to overlap:
			// an rpc that reaches the model because someone added it to both
			// lists is the failure this list exists to prevent.
			if escalation[full] {
				return nil, fmt.Errorf("tool: %s is excluded from the tools but also exposed", full)
			}
			op, ok := catalog.ByFullMethod(full)
			if !ok {
				return nil, fmt.Errorf("tool: %s is exposed but not in the API catalog", full)
			}
			name := toolName(op.Method)
			if previous, taken := box.byName[name]; taken {
				return nil, fmt.Errorf("tool: %s and %s both generate the tool name %q",
					previous.op.FullMethod, full, name)
			}
			box.byName[name] = binding{op: op, svc: svc, handler: methodHandler(method.Handler), write: write}
		}
	}
	return box, nil
}

// Tools lists what the user in ctx is allowed to call.
func (b *toolBox) Tools(ctx context.Context) ([]chatDomain.Tool, error) {
	userID, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("tool: no user in context")
	}

	names := make([]string, 0, len(b.byName))
	for name := range b.byName {
		names = append(names, name)
	}
	sort.Strings(names) // a stable tool order keeps the prompt cacheable

	tools := make([]chatDomain.Tool, 0, len(names))
	for _, name := range names {
		bind := b.byName[name]
		// Public means "any signed-in user", and the authorization interceptor
		// skips the check for it. Asking Enforce anyway would withhold a tool
		// the server would have run, and it withholds it from exactly the
		// people it is for: raising an access request is public because the
		// ones who most need to ask are the ones holding no permissions.
		if !bind.op.Public {
			allowed, err := b.authSvc.Enforce(ctx, auth.Request{
				Subject: userID,
				Object:  bind.op.Resource,
				Action:  bind.op.Action,
			})
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		tools = append(tools, chatDomain.Tool{
			Name:        name,
			Description: describe(bind.op, bind.write),
			InputSchema: schemaOf(bind.op),
			Write:       bind.write,
		})
	}
	return tools, nil
}

// Call dispatches one tool as the user in ctx.
func (b *toolBox) Call(ctx context.Context, name string, args []byte) (string, error) {
	bind, ok := b.byName[name]
	if !ok {
		return "", fmt.Errorf("no such tool: %s", name)
	}

	// The arguments come from a model, so they are decoded into the request
	// message with the same protojson the gateway uses and then validated by
	// the interceptor - not trusted as given.
	decode := func(req any) error {
		message, ok := req.(proto.Message)
		if !ok {
			return fmt.Errorf("tool: %s does not take a protobuf request", name)
		}
		if len(args) == 0 {
			return nil
		}
		return protojson.Unmarshal(args, message)
	}

	response, err := bind.handler(bind.svc.impl, ctx, decode, b.interceptor)
	if err != nil {
		return "", err
	}

	message, ok := response.(proto.Message)
	if !ok {
		return "", fmt.Errorf("tool: %s returned %T, not a protobuf message", name, response)
	}
	encoded, err := protojson.Marshal(message)
	if err != nil {
		return "", err
	}
	return redact(encoded)
}

// redact removes the hidden fields at any depth. It walks the decoded JSON
// rather than the message so that a field newly added to a response is covered
// without anyone remembering to update a per-message list.
func redact(encoded []byte) (string, error) {
	var body any
	if err := json.Unmarshal(encoded, &body); err != nil {
		return "", err
	}
	stripped, err := json.Marshal(strip(body))
	if err != nil {
		return "", err
	}
	return string(stripped), nil
}

func strip(node any) any {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if hidden[key] {
				delete(value, key)
				continue
			}
			value[key] = strip(child)
		}
		return value
	case []any:
		for i, child := range value {
			value[i] = strip(child)
		}
		return value
	default:
		return node
	}
}

// describe writes the tool description the model reads. The summary comes from
// the proto annotation, so it is the same sentence the CLI and the agent
// reference show.
func describe(op apicatalog.Operation, write bool) string {
	var b strings.Builder
	b.WriteString(op.Summary)
	if op.Summary != "" && !strings.HasSuffix(op.Summary, ".") {
		b.WriteString(".")
	}
	if write {
		// Said in the description so the model knows the pause is coming and
		// does not treat a proposal as a completed action.
		b.WriteString(" This changes data, and runs only after the user approves it.")
	} else {
		b.WriteString(" Read-only.")
	}
	return b.String()
}

// schemaOf renders the request message as a JSON Schema object.
func schemaOf(op apicatalog.Operation) map[string]any {
	properties := map[string]any{}
	var required []string

	for _, field := range op.Fields {
		property := map[string]any{}
		if description := fieldDescription(field); description != "" {
			property["description"] = description
		}
		if len(field.EnumValues) > 0 {
			property["enum"] = field.EnumValues
		}

		if field.Repeated {
			item := map[string]any{"type": jsonType(field.Kind)}
			if len(field.EnumValues) > 0 {
				item["enum"] = field.EnumValues
				delete(property, "enum")
			}
			property["type"] = "array"
			property["items"] = item
		} else {
			property["type"] = jsonType(field.Kind)
		}

		properties[field.JSONName] = property
		// A path parameter is part of the route, so the call cannot be made
		// without it.
		if field.In == apicatalog.InPath {
			required = append(required, field.JSONName)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		sort.Strings(required)
		schema["required"] = required
	}
	return schema
}

func fieldDescription(field apicatalog.Field) string {
	if len(field.Constraints) == 0 {
		return ""
	}
	return strings.Join(field.Constraints, ", ")
}

func jsonType(kind string) string {
	switch kind {
	case "bool":
		return "boolean"
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64":
		return "integer"
	case "float", "double":
		return "number"
	default:
		// enum, string, bytes and message all arrive as strings or objects the
		// model writes as text; string is the safe description.
		return "string"
	}
}

// toolName turns an rpc name into the identifier the model calls, which may
// only contain letters, digits, underscores and hyphens.
func toolName(method string) string {
	var b strings.Builder
	for i, r := range method {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
