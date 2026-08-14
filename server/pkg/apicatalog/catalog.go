// Package apicatalog turns the compiled protobuf descriptors into a single
// description of the HTTP/gRPC surface: what every rpc is called, how it is
// reached over REST, which fields it takes, and what RBAC demands of it.
//
// Three consumers used to answer those questions independently — the
// authorization interceptor parsed gRPC method strings, the CLI would have
// needed a hand-written command table, and the agent documentation was written
// by hand — so they drifted apart as the protos changed. They now all read this
// catalog, which is derived from the descriptors at startup and therefore
// cannot fall behind them.
package apicatalog

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	validate "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"

	hubv1 "github.com/linzhengen/hub/server/pb/hub/annotations/v1"

	// Importing the generated packages registers their descriptors with
	// protoregistry.GlobalFiles, which is what the catalog is built from. They
	// are blank imports because only the registration side effect is wanted.
	//
	// A service missing from this list is not a compile error and not a runtime
	// error: it silently loses its RBAC rule, its CLI command, its row in the
	// web client's operation table and its entry in the agent reference.
	// TestDefaultCoversEveryService reads proto/ and fails when one is missing,
	// so add the import here when you add a service.
	_ "github.com/linzhengen/hub/server/pb/ai/chat/v1"
	_ "github.com/linzhengen/hub/server/pb/system/audit/v1"
	_ "github.com/linzhengen/hub/server/pb/system/group/v1"
	_ "github.com/linzhengen/hub/server/pb/system/permission/v1"
	_ "github.com/linzhengen/hub/server/pb/system/resource/v1"
	_ "github.com/linzhengen/hub/server/pb/system/role/v1"
	_ "github.com/linzhengen/hub/server/pb/user/v1"
)

// Location says where a request field travels in the REST mapping.
type Location string

const (
	// InPath is a field substituted into the URL template.
	InPath Location = "path"
	// InQuery is a field sent as a query string parameter.
	InQuery Location = "query"
	// InBody is a field sent in the JSON request body.
	InBody Location = "body"
)

// Field describes one field of an rpc's request message.
type Field struct {
	// Name is the protobuf field name, e.g. "group_ids".
	Name string
	// JSONName is the name accepted over REST, e.g. "groupIds".
	JSONName string
	// Kind is the protobuf kind, e.g. "string", "bool", "uint32", "enum".
	Kind string
	// Repeated is true for list fields.
	Repeated bool
	// Optional is true for explicitly optional (proto3 `optional`) fields.
	Optional bool
	// EnumValues lists the accepted names when Kind is "enum".
	EnumValues []string
	// Message is the fully qualified type name when Kind is "message".
	Message string
	// In says where the field travels in the REST mapping.
	In Location
	// Constraints describes the buf.validate rules the server enforces, in the
	// short forms `hub api describe` prints: "required", "email", "uuid",
	// "len 1..64", "<= 200". Empty when the field is unconstrained.
	//
	// It is rendered here rather than left as the raw rule so a caller - an
	// agent, most of all - can read what will be accepted without knowing
	// protovalidate.
	Constraints []string
}

// Operation is a single rpc together with its REST mapping and RBAC rule.
type Operation struct {
	// Service is the fully qualified service name, e.g. "user.v1.UserService".
	Service string
	// ServiceName is the bare service name, e.g. "UserService".
	ServiceName string
	// Method is the rpc name, e.g. "ListUser".
	Method string
	// FullMethod is the gRPC method path, e.g. "/user.v1.UserService/ListUser".
	FullMethod string
	// HTTPMethod is the REST verb, e.g. "GET". Empty when the rpc has no REST mapping.
	HTTPMethod string
	// Path is the REST path template, e.g. "/api/v1/users/{id}".
	Path string
	// Body is the google.api.http body selector: "*", a field name, or empty.
	Body string
	// Resource is the RBAC resource identifier enforced on this rpc.
	Resource string
	// Action is the RBAC verb enforced on this rpc.
	Action string
	// Public is true when the rpc skips RBAC enforcement. Authentication is
	// still required.
	Public bool
	// Summary is the one-line description from the proto annotation.
	Summary string
	// Fields describes the request message, in declaration order.
	Fields []Field
}

// PathParams returns the field names the path template substitutes, in the
// order they appear in the path.
func (o Operation) PathParams() []string {
	var params []string
	for _, f := range o.Fields {
		if f.In == InPath {
			params = append(params, f.Name)
		}
	}
	return params
}

// Field returns the request field with the given protobuf or JSON name.
func (o Operation) Field(name string) (Field, bool) {
	for _, f := range o.Fields {
		if f.Name == name || f.JSONName == name {
			return f, true
		}
	}
	return Field{}, false
}

// Catalog is an immutable index of the API surface.
type Catalog struct {
	operations []Operation
	byFull     map[string]Operation
}

var (
	defaultOnce    sync.Once
	defaultCatalog *Catalog
)

// Default returns the catalog built from the globally registered descriptors.
// It is built once and safe for concurrent use.
func Default() *Catalog {
	defaultOnce.Do(func() {
		defaultCatalog = New(protoregistry.GlobalFiles)
	})
	return defaultCatalog
}

// FileRanger is the subset of protoregistry.Files the catalog needs, so tests
// can build a catalog from a restricted set of descriptors.
type FileRanger interface {
	RangeFiles(func(protoreflect.FileDescriptor) bool)
}

// New builds a catalog from the descriptors in files. Only rpcs carrying a
// google.api.http or hub.annotations.v1.method_rule option are included: those
// are the ones this platform owns, which keeps infrastructure services such as
// gRPC health checking out of the catalog.
func New(files FileRanger) *Catalog {
	c := &Catalog{byFull: map[string]Operation{}}
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		services := fd.Services()
		for i := range services.Len() {
			sd := services.Get(i)
			methods := sd.Methods()
			for j := range methods.Len() {
				op, ok := newOperation(sd, methods.Get(j))
				if !ok {
					continue
				}
				c.operations = append(c.operations, op)
				c.byFull[op.FullMethod] = op
			}
		}
		return true
	})
	sort.Slice(c.operations, func(i, j int) bool {
		if c.operations[i].Service != c.operations[j].Service {
			return c.operations[i].Service < c.operations[j].Service
		}
		return c.operations[i].Method < c.operations[j].Method
	})
	return c
}

// Operations returns every operation, ordered by service then method.
func (c *Catalog) Operations() []Operation {
	return c.operations
}

// ByFullMethod looks an operation up by its gRPC method path, the form
// grpc.UnaryServerInfo reports.
func (c *Catalog) ByFullMethod(fullMethod string) (Operation, bool) {
	op, ok := c.byFull[fullMethod]
	return op, ok
}

// Services returns the fully qualified service names, in catalog order.
func (c *Catalog) Services() []string {
	var (
		names []string
		seen  = map[string]bool{}
	)
	for _, op := range c.operations {
		if !seen[op.Service] {
			seen[op.Service] = true
			names = append(names, op.Service)
		}
	}
	return names
}

// OperationsOf returns the operations of one fully qualified service.
func (c *Catalog) OperationsOf(service string) []Operation {
	var ops []Operation
	for _, op := range c.operations {
		if op.Service == service {
			ops = append(ops, op)
		}
	}
	return ops
}

// Find resolves an operation from a human-written reference. It accepts the
// gRPC method path ("/user.v1.UserService/ListUser"), the qualified form
// ("user.v1.UserService.ListUser"), the short form ("UserService.ListUser") and
// the bare method name ("ListUser") as long as that is unambiguous.
func (c *Catalog) Find(ref string) (Operation, error) {
	if op, ok := c.byFull[ref]; ok {
		return op, nil
	}
	var matches []Operation
	for _, op := range c.operations {
		switch ref {
		case op.Service + "." + op.Method,
			op.ServiceName + "." + op.Method,
			op.Method:
			matches = append(matches, op)
		}
	}
	switch len(matches) {
	case 0:
		return Operation{}, fmt.Errorf("unknown operation %q", ref)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, op := range matches {
			names = append(names, op.Service+"."+op.Method)
		}
		return Operation{}, fmt.Errorf("ambiguous operation %q, matches: %s", ref, strings.Join(names, ", "))
	}
}

func newOperation(sd protoreflect.ServiceDescriptor, md protoreflect.MethodDescriptor) (Operation, bool) {
	httpRule := httpRuleOf(md)
	rule := methodRuleOf(md)
	if httpRule == nil && rule == nil {
		return Operation{}, false
	}

	service := string(sd.FullName())
	op := Operation{
		Service:     service,
		ServiceName: string(sd.Name()),
		Method:      string(md.Name()),
		FullMethod:  "/" + service + "/" + string(md.Name()),
		Resource:    ResourceOf(service),
		Action:      string(md.Name()),
	}
	if httpRule != nil {
		op.HTTPMethod, op.Path = httpMethodAndPath(httpRule)
		op.Body = httpRule.GetBody()
	}
	if rule != nil {
		op.Public = rule.GetPublic()
		op.Summary = rule.GetSummary()
		if rule.GetResource() != "" {
			op.Resource = rule.GetResource()
		}
		if rule.GetAction() != "" {
			op.Action = rule.GetAction()
		}
	}
	op.Fields = fieldsOf(md.Input(), op.Path, op.Body)
	return op, true
}

// ResourceOf returns the RBAC resource identifier a service defaults to. It
// mirrors the identifier `cli resource-import` registers for the service.
func ResourceOf(service string) string {
	return "api." + service
}

func httpRuleOf(md protoreflect.MethodDescriptor) *annotations.HttpRule {
	opts, ok := md.Options().(*descriptorpb.MethodOptions)
	if !ok || opts == nil {
		return nil
	}
	rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return nil
	}
	return rule
}

func methodRuleOf(md protoreflect.MethodDescriptor) *hubv1.MethodRule {
	opts, ok := md.Options().(*descriptorpb.MethodOptions)
	if !ok || opts == nil {
		return nil
	}
	rule, ok := proto.GetExtension(opts, hubv1.E_MethodRule).(*hubv1.MethodRule)
	if !ok {
		return nil
	}
	return rule
}

func httpMethodAndPath(rule *annotations.HttpRule) (string, string) {
	switch pattern := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return "GET", pattern.Get
	case *annotations.HttpRule_Put:
		return "PUT", pattern.Put
	case *annotations.HttpRule_Post:
		return "POST", pattern.Post
	case *annotations.HttpRule_Delete:
		return "DELETE", pattern.Delete
	case *annotations.HttpRule_Patch:
		return "PATCH", pattern.Patch
	case *annotations.HttpRule_Custom:
		return strings.ToUpper(pattern.Custom.GetKind()), pattern.Custom.GetPath()
	default:
		return "", ""
	}
}

func fieldsOf(md protoreflect.MessageDescriptor, path, body string) []Field {
	pathParams := pathParamsOf(path)
	fds := md.Fields()
	fields := make([]Field, 0, fds.Len())
	for i := range fds.Len() {
		fd := fds.Get(i)
		f := Field{
			Name:     string(fd.Name()),
			JSONName: fd.JSONName(),
			Kind:     fd.Kind().String(),
			Repeated: fd.IsList(),
			Optional: fd.HasOptionalKeyword(),
			In:       locationOf(string(fd.Name()), pathParams, body),
		}
		f.Constraints = constraintsOf(fd)

		switch fd.Kind() {
		case protoreflect.EnumKind:
			values := fd.Enum().Values()
			for j := range values.Len() {
				f.EnumValues = append(f.EnumValues, string(values.Get(j).Name()))
			}
		case protoreflect.MessageKind, protoreflect.GroupKind:
			f.Message = string(fd.Message().FullName())
			if fd.IsMap() {
				f.Kind = "map"
			}
		}
		fields = append(fields, f)
	}
	return fields
}

// constraintsOf renders a field's buf.validate rules into short human phrases.
//
// Only the rules this codebase actually uses are translated; anything else is
// left out rather than half-described, so what is printed is always accurate.
func constraintsOf(fd protoreflect.FieldDescriptor) []string {
	opts, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil {
		return nil
	}
	rules, ok := proto.GetExtension(opts, validate.E_Field).(*validate.FieldRules)
	if !ok || rules == nil {
		return nil
	}

	var out []string
	if rules.GetRequired() {
		out = append(out, "required")
	}
	if repeated := rules.GetRepeated(); repeated != nil {
		if repeated.GetMinItems() > 0 {
			out = append(out, fmt.Sprintf("at least %d item(s)", repeated.GetMinItems()))
		}
		if repeated.GetMaxItems() > 0 {
			out = append(out, fmt.Sprintf("at most %d item(s)", repeated.GetMaxItems()))
		}
		out = append(out, prefixed("each ", stringConstraints(repeated.GetItems().GetString()))...)
	}
	out = append(out, stringConstraints(rules.GetString())...)
	if u := rules.GetUint32(); u != nil && u.GetLte() > 0 {
		out = append(out, fmt.Sprintf("<= %d", u.GetLte()))
	}
	if u := rules.GetUint64(); u != nil && u.GetLte() > 0 {
		out = append(out, fmt.Sprintf("<= %d", u.GetLte()))
	}
	return out
}

func stringConstraints(rules *validate.StringRules) []string {
	if rules == nil {
		return nil
	}

	var out []string
	switch rules.GetWellKnown().(type) {
	case *validate.StringRules_Email:
		out = append(out, "email")
	case *validate.StringRules_Uuid:
		out = append(out, "uuid")
	case *validate.StringRules_Uri:
		out = append(out, "uri")
	}
	switch {
	case rules.MinLen != nil && rules.MaxLen != nil:
		out = append(out, fmt.Sprintf("length %d..%d", rules.GetMinLen(), rules.GetMaxLen()))
	case rules.MinLen != nil:
		out = append(out, fmt.Sprintf("length >= %d", rules.GetMinLen()))
	case rules.MaxLen != nil:
		out = append(out, fmt.Sprintf("length <= %d", rules.GetMaxLen()))
	}
	return out
}

func prefixed(prefix string, values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, prefix+v)
	}
	return out
}

func locationOf(name string, pathParams map[string]bool, body string) Location {
	switch {
	case pathParams[name]:
		return InPath
	case body == "*", body == name:
		return InBody
	default:
		return InQuery
	}
}

// pathParamsOf extracts the field names a google.api.http path template
// substitutes. Templates take the forms "{name}" and "{name=pattern}".
func pathParamsOf(path string) map[string]bool {
	params := map[string]bool{}
	for rest := path; ; {
		open := strings.Index(rest, "{")
		if open < 0 {
			return params
		}
		rest = rest[open+1:]
		end := strings.Index(rest, "}")
		if end < 0 {
			return params
		}
		name, _, _ := strings.Cut(rest[:end], "=")
		params[name] = true
		rest = rest[end+1:]
	}
}
