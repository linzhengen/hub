package register

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sethvargo/go-limiter/memorystore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/oidc/token"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/interface/grpc/interceptor"
	pbchatv1 "github.com/linzhengen/hub/server/pb/ai/chat/v1"
	pbaccessv1 "github.com/linzhengen/hub/server/pb/system/access/v1"
	pbauditv1 "github.com/linzhengen/hub/server/pb/system/audit/v1"
	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	pborganizationv1 "github.com/linzhengen/hub/server/pb/system/organization/v1"
	pbpermissionv1 "github.com/linzhengen/hub/server/pb/system/permission/v1"
	pbresourcev1 "github.com/linzhengen/hub/server/pb/system/resource/v1"
	pbrolev1 "github.com/linzhengen/hub/server/pb/system/role/v1"
	pbserviceaccountv1 "github.com/linzhengen/hub/server/pb/system/serviceaccount/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
	"github.com/linzhengen/hub/server/pkg/logger"
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
	auditRepo audit.Repository,
	transRepo trans.Repository,
	catalog *apicatalog.Catalog,
	validator interceptor.Validator,
	roleServiceServer pbrolev1.RoleServiceServer,
	userServiceServer pbuserv1.UserServiceServer,
	permissionServiceServer pbpermissionv1.PermissionServiceServer,
	resourceServiceServer pbresourcev1.ResourceServiceServer,
	groupServiceServer pbgroupv1.GroupServiceServer,
	organizationServiceServer pborganizationv1.OrganizationServiceServer,
	chatServiceServer pbchatv1.ChatServiceServer,
	auditServiceServer pbauditv1.AuditServiceServer,
	accessServiceServer pbaccessv1.AccessServiceServer,
	accessRequestServiceServer pbaccessv1.AccessRequestServiceServer,
	serviceAccountServiceServer pbserviceaccountv1.ServiceAccountServiceServer,
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
			// Before authorization, so that a refused attempt is recorded too.
			interceptor.UnaryAuditInterceptor(auditRepo, transRepo, catalog),
			interceptor.UnaryAuthzInterceptor(authSvc, userRepo, catalog),
			interceptor.UnaryValidateInterceptor(validator),
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
			interceptor.StreamAuthzInterceptor(authSvc, userRepo, catalog),
			interceptor.StreamValidateInterceptor(validator),
			interceptor.ErrorTranslationStreamServerInterceptor,
			//s.gatekeeper.StreamServerInterceptor(),
			interceptor.RatelimitStreamServerInterceptor(store),
			//interceptor.SetVersionHeaderStreamServerInterceptor(opts.Version),
		)),
	}

	// A mutating rpc nobody classified would go unrecorded, so say so here as
	// well as in CI: a deployment running ahead of the classification is the
	// case the test cannot catch.
	if missing := interceptor.UnclassifiedMutations(catalog); len(missing) > 0 {
		logger.Errorf("audit: these rpcs change state but are not classified as audited or not: %v", missing)
	}

	grpcServer := grpc.NewServer(sOpts...)
	//grpc_prometheus.Register(grpcServer)

	healthServer := health.NewServer()
	pbuserv1.RegisterUserServiceServer(grpcServer, userServiceServer)
	pbrolev1.RegisterRoleServiceServer(grpcServer, roleServiceServer)
	pbpermissionv1.RegisterPermissionServiceServer(grpcServer, permissionServiceServer)
	pbresourcev1.RegisterResourceServiceServer(grpcServer, resourceServiceServer)
	pbgroupv1.RegisterGroupServiceServer(grpcServer, groupServiceServer)
	pborganizationv1.RegisterOrganizationServiceServer(grpcServer, organizationServiceServer)
	pbchatv1.RegisterChatServiceServer(grpcServer, chatServiceServer)
	pbauditv1.RegisterAuditServiceServer(grpcServer, auditServiceServer)
	pbaccessv1.RegisterAccessServiceServer(grpcServer, accessServiceServer)
	pbaccessv1.RegisterAccessRequestServiceServer(grpcServer, accessRequestServiceServer)
	pbserviceaccountv1.RegisterServiceAccountServiceServer(grpcServer, serviceAccountServiceServer)
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	return grpcServer
}
