package interceptor

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Validator checks a request against the constraints its .proto declares.
//
// It is protovalidate's own interface rather than a narrowed one, so the
// interceptors can be handed a real validator without an adapter and a test
// can substitute a stub.
type Validator = protovalidate.Validator

// NewValidator builds the protovalidate validator the interceptors use. It
// compiles every constraint in the registered descriptors up front, so a
// malformed rule is a startup failure rather than a surprise on the first
// request that reaches it.
func NewValidator() (Validator, error) {
	return protovalidate.New()
}

// UnaryValidateInterceptor rejects a request that breaks its own constraints.
//
// Validation used to be wherever a handler or use case happened to remember it,
// which meant an empty username or a malformed email reached Keycloak and a
// list request could ask for any number of rows. Enforcing the proto's rules in
// one place makes the definition the contract: what the .proto says is what the
// server accepts, and `hub api describe` can show a caller the same rules
// before they send anything.
//
// It runs after authorization, so an unauthorised caller cannot use the
// difference between InvalidArgument and PermissionDenied to probe the shape of
// an endpoint they may not reach.
func UnaryValidateInterceptor(validator Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validateMessage(validator, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamValidateInterceptor validates each message a client streams.
func StreamValidateInterceptor(validator Validator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &validatingServerStream{ServerStream: stream, validator: validator})
	}
}

// validateMessage returns an InvalidArgument status describing what is wrong,
// or nil. A request that is not a protobuf message is passed through: there is
// nothing to check it against.
func validateMessage(validator Validator, req interface{}) error {
	message, ok := req.(proto.Message)
	if !ok {
		return nil
	}
	if err := validator.Validate(message); err != nil {
		// protovalidate's message names the offending fields and why, which is
		// what the caller needs in order to fix the request.
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

// validatingServerStream checks messages as they are received rather than all
// at once: a stream has no single request to validate.
type validatingServerStream struct {
	grpc.ServerStream
	validator Validator
}

func (s *validatingServerStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	return validateMessage(s.validator, m)
}
