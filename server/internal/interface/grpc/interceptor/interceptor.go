package interceptor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sethvargo/go-limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/oidc/token"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/group"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

// PanicLoggerUnaryServerInterceptor returns a new unary server interceptor for recovering from panics and returning error
func PanicLoggerUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Info(fmt.Sprintf("Recovered from panic: %+v\n%s", r, debug.Stack()))
				err = status.Errorf(codes.Internal, "%s", r)
			}
		}()
		return handler(ctx, req)
	}
}

// PanicLoggerStreamServerInterceptor returns a new streaming server interceptor for recovering from panics and returning error
func PanicLoggerStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Info(fmt.Sprintf("Recovered from panic: %+v\n%s", r, debug.Stack()))
				err = status.Errorf(codes.Internal, "%s", r)
			}
		}()
		return handler(srv, stream)
	}
}

const (
	HubVersionHeader = "hub-version"
)

var (
	LastSeenServerVersion                  string
	ErrorTranslationUnaryServerInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		return resp, TranslateError(err)
	}
	ErrorTranslationStreamServerInterceptor = func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return TranslateError(handler(srv, ss))
	}
)

// SetVersionHeaderUnaryServerInterceptor returns a new unary server interceptor that sets the argo-version header
func SetVersionHeaderUnaryServerInterceptor(version string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		m, origErr := handler(ctx, req)
		if origErr == nil {
			// Don't set header if there was an error because attackers could use it to find vulnerable Argo servers
			err := grpc.SetHeader(ctx, metadata.Pairs(HubVersionHeader, version))
			if err != nil {
				logger.Errorf("Failed to set header '%s': %s", HubVersionHeader, err)
			}
		}
		return m, origErr
	}
}

// SetVersionHeaderStreamServerInterceptor returns a new stream server interceptor that sets the argo-version header
func SetVersionHeaderStreamServerInterceptor(version string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		origErr := handler(srv, ss)
		if origErr == nil {
			// Don't set header if there was an error because attackers could use it to find vulnerable Argo servers
			err := ss.SetHeader(metadata.Pairs(HubVersionHeader, version))
			if err != nil {
				logger.Errorf("Failed to set header '%s': %s", HubVersionHeader, err)
			}
		}
		return origErr
	}
}

// GetVersionHeaderClientUnaryInterceptor returns a new unary client interceptor that extracts the argo-version from the response and sets the global variable LastSeenServerVersion
func GetVersionHeaderClientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var headers metadata.MD
	err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Header(&headers))...)
	if err == nil && headers != nil && headers.Get(HubVersionHeader) != nil {
		LastSeenServerVersion = headers.Get(HubVersionHeader)[0]
	}
	return err
}

// RatelimitUnaryServerInterceptor returns a new unary server interceptor that performs request rate limiting.
func RatelimitUnaryServerInterceptor(ratelimiter limiter.Store) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ip := getClientIP(ctx)
		_, _, _, ok, err := ratelimiter.Take(ctx, ip)
		if err != nil {
			logger.Errorf("Internal Server Error: %s", err)
			return nil, status.Errorf(codes.Internal, "%s: grpc_ratelimit middleware internal error", info.FullMethod)
		}
		if !ok {
			return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
		}
		return handler(ctx, req)
	}
}

// RatelimitStreamServerInterceptor returns a new stream server interceptor that performs rate limiting on the request.
func RatelimitStreamServerInterceptor(ratelimiter limiter.Store) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		ip := getClientIP(ctx)
		_, _, _, ok, err := ratelimiter.Take(ctx, ip)
		if err != nil {
			logger.Errorf("Internal Server Error: %s", err)
			return status.Errorf(codes.Internal, "%s: grpc_ratelimit middleware internal error", info.FullMethod)
		}
		if !ok {
			return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
		}
		return handler(srv, stream)
	}
}

// GetClientIP inspects the context to retrieve the ip address of the client
func getClientIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Errorf("couldn't parse client IP address")
		return ""
	}
	address := p.Addr.String()
	ip := strings.Split(address, ":")[0]
	return ip
}

func UnaryAuthInterceptor(tokenOpe token.Operator, userSvc user.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		t, err := tokenOpe.ExtractToken(ctx)
		if err != nil {
			return ctx, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		if err := userSvc.CreateIfNotExists(ctx, userFactory(t)); err != nil {
			return ctx, status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}

		return handler(contextx.WithUserID(ctx, t.UserId), req)
	}
}

func StreamAuthInterceptor(tokenOpe token.Operator, userSvc user.Service) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		t, err := tokenOpe.ExtractToken(ctx)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		if err := userSvc.CreateIfNotExists(ctx, userFactory(t)); err != nil {
			return status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}
		return handler(srv, &authServerStream{ctx, t.UserId, stream})
	}
}

func userFactory(t *token.Token) *user.User {
	var groupIds []string
	if len(t.Roles) > 0 {
		for _, r := range t.Roles {
			if r == token.AdminRoleName {
				groupIds = append(groupIds, group.AdminGroupId)
				break
			}
		}
	}
	return &user.User{
		Id:        t.UserId,
		Username:  t.Username,
		Email:     t.Email,
		Status:    user.Active,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),

		GroupIds: groupIds,
	}
}

type authServerStream struct {
	ctx    context.Context
	userId string
	grpc.ServerStream
}

func (a *authServerStream) Context() context.Context {
	return a.ctx
}

func (a *authServerStream) RecvMsg(m interface{}) error {
	return a.ServerStream.RecvMsg(m)
}

func (a *authServerStream) SendMsg(m interface{}) error {
	err := a.ServerStream.SendMsg(m)
	if err != nil {
		return err
	}
	a.ctx = contextx.WithUserID(a.ctx, a.userId)
	return nil
}

// authzUserLookupError maps a userRepo.FindOne error to a gRPC status: a
// missing user (sql.ErrNoRows) is a genuine authorization failure, while any
// other error (e.g. a DB timeout) is an infrastructure failure and must not
// be reported to the client as PermissionDenied.
func authzUserLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return status.Errorf(codes.PermissionDenied, "permission denied: invalid user")
	}
	return status.Errorf(codes.Internal, "failed to look up user: %v", err)
}

// UnaryAuthzInterceptor creates a new unary server interceptor for authorization
func UnaryAuthzInterceptor(authSvc auth.Service, userRepo user.Repository, catalog *apicatalog.Catalog) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := authorize(ctx, authSvc, userRepo, catalog, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamAuthzInterceptor creates a new stream server interceptor for authorization
func StreamAuthzInterceptor(authSvc auth.Service, userRepo user.Repository, catalog *apicatalog.Catalog) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorize(stream.Context(), authSvc, userRepo, catalog, info.FullMethod); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

// authorize runs the checks both authz interceptors share: the caller is known,
// their account is still active, and the rpc's RBAC rule permits the call. It
// returns a gRPC status error, or nil when the call may proceed.
func authorize(
	ctx context.Context,
	authSvc auth.Service,
	userRepo user.Repository,
	catalog *apicatalog.Catalog,
	fullMethod string,
) error {
	userID, ok := contextx.GetUserID(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	u, err := userRepo.FindOne(ctx, userID)
	if err != nil {
		return authzUserLookupError(err)
	}
	if u.Status == user.InActive {
		return status.Errorf(codes.PermissionDenied, "permission denied: user is inactive")
	}

	resource, action, public := authzTarget(catalog, fullMethod)
	if public {
		return nil
	}

	allowed, err := authSvc.Enforce(ctx, auth.Request{
		Subject: userID,
		Object:  resource,
		Action:  action,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "authorization error: %v", err)
	}
	if !allowed {
		return status.Errorf(codes.PermissionDenied, "permission denied for %s on %s", action, resource)
	}
	return nil
}

// authzTarget resolves what an rpc is enforced against. The answer comes from
// the rpc's hub.annotations.v1.method_rule option by way of the API catalog, so
// making an endpoint public is a one-line proto change rather than an edit to
// this file. Methods outside the catalog - gRPC health checking, say - keep
// falling back to the naming convention the catalog itself defaults to.
func authzTarget(catalog *apicatalog.Catalog, fullMethod string) (resource, action string, public bool) {
	if catalog != nil {
		if op, ok := catalog.ByFullMethod(fullMethod); ok {
			return op.Resource, op.Action, op.Public
		}
	}
	resource, action = extractResourceAndAction(fullMethod)
	return resource, action, false
}

// extractResourceAndAction extracts the resource and action from the gRPC method
// Format: /package.service/method
func extractResourceAndAction(fullMethod string) (string, string) {
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 3 {
		return "", ""
	}

	service := parts[1]
	action := parts[2]
	return apicatalog.ResourceOf(service), action
}
