package register

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sethvargo/go-limiter/memorystore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
	"github.com/linzhengen/hub/v1/server/internal/domain/oidc/token"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/internal/interface/grpc/interceptor"
	pbgroupv1 "github.com/linzhengen/hub/v1/server/pb/system/group/v1"
	pbpermissionv1 "github.com/linzhengen/hub/v1/server/pb/system/permission/v1"
	pbresourcev1 "github.com/linzhengen/hub/v1/server/pb/system/resource/v1"
	pbrolev1 "github.com/linzhengen/hub/v1/server/pb/system/role/v1"
	pbuserv1 "github.com/linzhengen/hub/v1/server/pb/user/v1"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type Opts struct {
	APIRateLimit       uint64
	MaxGRPCMessageSize int
	Version            string
	DisableAuth        bool
}

func New(
	opts *Opts,
	tokenOpe token.Operator,
	userSvc user.Service,
	userRepo user.Repository,
	authSvc auth.Service,
	roleServiceServer pbrolev1.RoleServiceServer,
	userServiceServer pbuserv1.UserServiceServer,
	permissionServiceServer pbpermissionv1.PermissionServiceServer,
	resourceServiceServer pbresourcev1.ResourceServiceServer,
	groupServiceServer pbgroupv1.GroupServiceServer,
) *grpc.Server {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   opts.APIRateLimit,
		Interval: time.Second,
	})
	if err != nil {
		logger.Severef("failed to create rate limiter store: %v", err)
	}
	//grpc_prometheus.EnableHandlingTimeHistogram()
	sOpts := []grpc.ServerOption{
		// Set both the send and receive the bytes limit to be 100MB or GRPC_MESSAGE_SIZE
		// The proper way to achieve high performance is to have pagination
		// while we work toward that, we can have high limit first
		grpc.MaxRecvMsgSize(opts.MaxGRPCMessageSize),
		grpc.MaxSendMsgSize(opts.MaxGRPCMessageSize),
		grpc.ConnectionTimeout(300 * time.Second),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			//grpc_prometheus.UnaryServerInterceptor,
			//grpc_zap.UnaryServerInterceptor(log),
			interceptor.PanicLoggerUnaryServerInterceptor(),
			interceptor.LoggingUnaryServerInterceptor(),
			interceptor.UnaryAuthInterceptor(tokenOpe, userSvc),
			interceptor.UnaryAuthzInterceptor(authSvc, userRepo),
			interceptor.ErrorTranslationUnaryServerInterceptor,
			interceptor.RatelimitUnaryServerInterceptor(store),
			//interceptor.SetVersionHeaderUnaryServerInterceptor(opts.Version),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			//grpc_prometheus.StreamServerInterceptor,
			//grpc_zap.StreamServerInterceptor(serverLog),
			interceptor.PanicLoggerStreamServerInterceptor(),
			interceptor.LoggingStreamServerInterceptor(),
			interceptor.StreamAuthInterceptor(tokenOpe, userSvc),
			interceptor.StreamAuthzInterceptor(authSvc, userRepo),
			interceptor.ErrorTranslationStreamServerInterceptor,
			//s.gatekeeper.StreamServerInterceptor(),
			interceptor.RatelimitStreamServerInterceptor(store),
			//interceptor.SetVersionHeaderStreamServerInterceptor(opts.Version),
		)),
	}

	grpcServer := grpc.NewServer(sOpts...)
	//grpc_prometheus.Register(grpcServer)

	healthServer := health.NewServer()
	pbuserv1.RegisterUserServiceServer(grpcServer, userServiceServer)
	pbrolev1.RegisterRoleServiceServer(grpcServer, roleServiceServer)
	pbpermissionv1.RegisterPermissionServiceServer(grpcServer, permissionServiceServer)
	pbresourcev1.RegisterResourceServiceServer(grpcServer, resourceServiceServer)
	pbgroupv1.RegisterGroupServiceServer(grpcServer, groupServiceServer)
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	return grpcServer
}
