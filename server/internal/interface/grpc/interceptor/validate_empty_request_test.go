package interceptor_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/linzhengen/hub/server/internal/interface/grpc/interceptor"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// TestUnfilteredListingsValidate checks that every rpc a caller can reach with
// no arguments accepts having none.
//
// A GET with no path parameters is a listing: every field on it is a filter, and
// a caller who wants all the rows sends nothing. protovalidate does not know
// that. A rule on a bare `string` field runs even when the field is absent, so
// `org_id` declared as
//
//	string org_id = 7 [(buf.validate.field).string.uuid = true];
//
// refuses every unfiltered listing with "value is empty, which is not a valid
// UUID" - which is exactly what shipped, on two rpcs at once, until this test
// existed. Declaring the field `optional` skips the rule when it is absent,
// which is the pattern the rest of the API already follows.
//
// Marking a filter required is therefore a deliberate act, and the list below is
// where it is made.
//
// These are GETs whose arguments are the question rather than a filter on one:
// "why may this person do that" has no unfiltered form to fall back to. They are
// named here, with the reason, so that adding a third is a decision somebody
// takes rather than a rule that quietly stops guarding anything.
var requiresArguments = map[string]string{
	"/system.access.v1.AccessService/ExplainUserAccess": "the user, resource and action are the question: " +
		"there is no such thing as explaining nobody's access to nothing",
	"/system.access.v1.AccessService/ListPrincipalsForOperation": "the resource and action are the question: " +
		"it answers who may do one particular thing",
}

func TestUnfilteredListingsValidate(t *testing.T) {
	validator, err := interceptor.NewValidator()
	require.NoError(t, err)

	for _, op := range apicatalog.Default().Operations() {
		if op.HTTPMethod != "GET" || len(op.PathParams()) > 0 {
			continue
		}
		if _, exempt := requiresArguments[op.FullMethod]; exempt {
			continue
		}

		t.Run(op.FullMethod, func(t *testing.T) {
			service, method, ok := strings.Cut(strings.TrimPrefix(op.FullMethod, "/"), "/")
			require.True(t, ok, "unexpected method path %q", op.FullMethod)

			desc, err := protoregistry.GlobalFiles.FindDescriptorByName(
				protoreflect.FullName(service + "." + method))
			require.NoError(t, err)

			md, ok := desc.(protoreflect.MethodDescriptor)
			require.True(t, ok, "%s is not a method descriptor", op.FullMethod)

			mt, err := protoregistry.GlobalTypes.FindMessageByName(md.Input().FullName())
			require.NoError(t, err)

			require.NoError(t, validator.Validate(mt.New().Interface()),
				"%s refuses a request with no filters set; declare its optional "+
					"filters as `optional` so the rule is skipped when they are absent",
				op.FullMethod)
		})
	}
}

// TestRequiresArgumentsIsCurrent keeps the exemption list honest: an entry that
// no longer refuses an empty request is an entry that has stopped documenting
// anything, and would hide the next rpc that starts refusing one.
func TestRequiresArgumentsIsCurrent(t *testing.T) {
	validator, err := interceptor.NewValidator()
	require.NoError(t, err)

	catalog := apicatalog.Default()
	for fullMethod, reason := range requiresArguments {
		op, ok := catalog.ByFullMethod(fullMethod)
		require.True(t, ok, "%s is not in the catalog any more", fullMethod)
		require.NotEmpty(t, reason, "%s needs a reason", fullMethod)

		service, method, ok := strings.Cut(strings.TrimPrefix(op.FullMethod, "/"), "/")
		require.True(t, ok)
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(
			protoreflect.FullName(service + "." + method))
		require.NoError(t, err)
		md, ok := desc.(protoreflect.MethodDescriptor)
		require.True(t, ok)
		mt, err := protoregistry.GlobalTypes.FindMessageByName(md.Input().FullName())
		require.NoError(t, err)

		require.Error(t, validator.Validate(mt.New().Interface()),
			"%s accepts an empty request now, so remove it from requiresArguments", fullMethod)
	}
}
