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

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/oidc/token"
	"github.com/linzhengen/hub/server/internal/domain/system/group"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
	"github.com/linzhengen/hub/server/pkg/logger"
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
	// HubOrgHeader names the organization a request is about.
	//
	// It is a header rather than a field on every request message because it
	// qualifies the question rather than being part of it: the same ListUser
	// asked in two organizations is the same rpc. Absent means "anywhere I hold
	// access", which is what every caller meant before organizations existed,
	// so a client that has never heard of this header keeps working.
	//
	// grpc-gateway forwards an HTTP header as `grpcgateway-<name>` unless it is
	// permitted, and lowercases it either way; both spellings are read.
	HubOrgHeader = "hub-org"
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
		u, err := userSvc.CreateIfNotExists(ctx, userFactory(t))
		if err != nil {
			return ctx, status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}

		return handler(authenticated(ctx, t.UserId, u), req)
	}
}

func StreamAuthInterceptor(tokenOpe token.Operator, userSvc user.Service) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		t, err := tokenOpe.ExtractToken(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		u, err := userSvc.CreateIfNotExists(stream.Context(), userFactory(t))
		if err != nil {
			return status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}
		return handler(srv, &authServerStream{
			ctx:          authenticated(stream.Context(), t.UserId, u),
			ServerStream: stream,
		})
	}
}

// authenticated records who the caller is, the row proving they still exist,
// and which organization they are asking about, for the rest of the request to
// read.
func authenticated(ctx context.Context, userID string, u *user.User) context.Context {
	ctx = contextx.WithOrgID(contextx.WithUserID(ctx, userID), requestedOrg(ctx))
	return user.NewContext(ctx, u)
}

// requestedOrg reads the organization off the request metadata, empty when the
// caller named none.
//
// Nothing is inferred when it is absent. hub could pick one - the caller's only
// organization, say - but a default that changes as somebody's memberships
// change is a decision that silently changes with them; a caller who cares
// which organization they are acting in has to say so.
func requestedOrg(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{HubOrgHeader, "grpcgateway-" + HubOrgHeader} {
		if values := md.Get(key); len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
	}
	return ""
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

		Groups: user.PermanentMemberships(groupIds),
	}
}

// authServerStream replaces a stream's context with the authenticated one, so
// that the interceptors and handlers downstream see the caller.
//
// It used to hand on the unauthenticated context and only add the user id
// after the first SendMsg, which is backwards: the authorization interceptor
// runs before any message is sent, so it saw no user id and rejected every
// streaming call as unauthenticated. Nothing streams today, so the fault was
// never reached.
type authServerStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (a *authServerStream) Context() context.Context {
	return a.ctx
}

// authenticatedUser returns the caller's row, preferring the one the
// authentication interceptor already read. The fallback query is for callers
// that reach authorization without having gone through it - a direct unit
// test, or a future interceptor chain that puts them in a different order.
func authenticatedUser(ctx context.Context, userRepo user.Repository, userID string) (*user.User, error) {
	if u, ok := user.FromContext(ctx); ok {
		return u, nil
	}
	u, err := userRepo.FindOne(ctx, userID)
	if err != nil {
		return nil, authzUserLookupError(err)
	}
	return u, nil
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

	u, err := authenticatedUser(ctx, userRepo, userID)
	if err != nil {
		return err
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
		OrgId:   contextx.GetOrgID(ctx),
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
