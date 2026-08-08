package interceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
)

func newTestValidator(t *testing.T) Validator {
	t.Helper()
	validator, err := NewValidator()
	require.NoError(t, err)
	return validator
}

// A malformed rule anywhere in the registered descriptors is a startup
// failure, not a surprise on the first request that reaches it.
func TestNewValidatorCompilesEveryRegisteredConstraint(t *testing.T) {
	validator, err := NewValidator()

	require.NoError(t, err)
	assert.NotNil(t, validator)
}

func TestUnaryValidateInterceptor(t *testing.T) {
	interceptor := UnaryValidateInterceptor(newTestValidator(t))
	info := &grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/CreateUser"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "handled", nil }

	valid := &pbuserv1.CreateUserRequest{
		Username: "taro",
		Email:    "taro@example.com",
		Password: "correct horse",
		GroupIds: []string{"00000000-0000-0000-0000-000000000001"},
	}

	t.Run("a valid request reaches the handler", func(t *testing.T) {
		resp, err := interceptor(context.Background(), valid, info, handler)

		require.NoError(t, err)
		assert.Equal(t, "handled", resp)
	})

	tests := []struct {
		name     string
		request  *pbuserv1.CreateUserRequest
		contains string
	}{
		{
			name:     "an empty username",
			request:  &pbuserv1.CreateUserRequest{Username: "", Email: "taro@example.com", Password: "correct horse"},
			contains: "username",
		},
		{
			// Previously this reached Keycloak before anything objected.
			name:     "a malformed email",
			request:  &pbuserv1.CreateUserRequest{Username: "taro", Email: "not-an-email", Password: "correct horse"},
			contains: "email",
		},
		{
			name:     "a short password",
			request:  &pbuserv1.CreateUserRequest{Username: "taro", Email: "taro@example.com", Password: "short"},
			contains: "password",
		},
		{
			name: "a group id that is not a uuid",
			request: &pbuserv1.CreateUserRequest{
				Username: "taro", Email: "taro@example.com", Password: "correct horse",
				GroupIds: []string{"not-a-uuid"},
			},
			contains: "group_ids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			resp, err := interceptor(context.Background(), tt.request, info,
				func(ctx context.Context, req interface{}) (interface{}, error) {
					handlerCalled = true
					return "handled", nil
				})

			assert.Nil(t, resp)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tt.contains, "the error should name the offending field")
			assert.False(t, handlerCalled, "the handler must not run on a rejected request")
		})
	}
}

// A list endpoint used to accept any limit and silently return the capped
// page. Asking for more than the server will ever give is now an error.
func TestUnaryValidateInterceptorBoundsListLimits(t *testing.T) {
	interceptor := UnaryValidateInterceptor(newTestValidator(t))
	info := &grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/ListUser"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "handled", nil }

	_, err := interceptor(context.Background(), &pbuserv1.ListUserRequest{Limit: 200}, info, handler)
	assert.NoError(t, err, "the maximum page is allowed")

	_, err = interceptor(context.Background(), &pbuserv1.ListUserRequest{Limit: 201}, info, handler)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// A link request with no ids does nothing, so it is a mistake rather than a
// no-op worth accepting.
func TestUnaryValidateInterceptorRejectsAnEmptyLinkRequest(t *testing.T) {
	interceptor := UnaryValidateInterceptor(newTestValidator(t))
	info := &grpc.UnaryServerInfo{FullMethod: "/system.group.v1.GroupService/AddRolesToGroup"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "handled", nil }

	id := "00000000-0000-0000-0000-000000000001"

	_, err := interceptor(context.Background(), &pbgroupv1.AddRolesToGroupRequest{Id: id}, info, handler)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = interceptor(context.Background(),
		&pbgroupv1.AddRolesToGroupRequest{Id: id, RoleIds: []string{id}}, info, handler)
	assert.NoError(t, err)
}

// Nothing streams today, but the interceptor is registered, so it must not
// reject or drop what it cannot check.
func TestValidateMessagePassesThroughANonProtoRequest(t *testing.T) {
	assert.NoError(t, validateMessage(newTestValidator(t), "not a proto message"))
}

func TestStreamValidateInterceptorChecksEachMessage(t *testing.T) {
	validator := newTestValidator(t)
	interceptor := StreamValidateInterceptor(validator)

	received := &pbuserv1.CreateUserRequest{Username: "", Email: "bad", Password: "x"}
	stream := &recvStream{ctx: context.Background(), message: received}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, func(_ interface{}, s grpc.ServerStream) error {
		return s.RecvMsg(&pbuserv1.CreateUserRequest{})
	})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// recvStream hands the handler one prepared message.
type recvStream struct {
	grpc.ServerStream
	ctx     context.Context
	message *pbuserv1.CreateUserRequest
}

func (s *recvStream) Context() context.Context { return s.ctx }

func (s *recvStream) RecvMsg(m interface{}) error {
	target, ok := m.(*pbuserv1.CreateUserRequest)
	if !ok {
		return nil
	}
	// proto.Merge rather than assignment: a message value carries a mutex.
	proto.Merge(target, s.message)
	return nil
}
