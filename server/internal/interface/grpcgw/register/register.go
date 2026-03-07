package register

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthv1pb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/linzhengen/hub/server/config"
	pbgrouupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	pbpermissionv1 "github.com/linzhengen/hub/server/pb/system/permission/v1"
	pbresourcev1 "github.com/linzhengen/hub/server/pb/system/resource/v1"
	pbrolev1 "github.com/linzhengen/hub/server/pb/system/role/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"

	"github.com/linzhengen/hub/server/pkg/logger"
)

func New(
	envCfg config.EnvConfig,
) *runtime.ServeMux {
	ctx := context.Background()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.NewClient(envCfg.Addr(), opts...)
	if err != nil {
		logger.Severe(err)
	}
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithHealthzEndpoint(healthv1pb.NewHealthClient(conn)),
	}
	mux := runtime.NewServeMux(muxOpts...)
	must(pbuserv1.RegisterUserServiceHandler(ctx, mux, conn))
	must(pbrolev1.RegisterRoleServiceHandler(ctx, mux, conn))
	must(pbpermissionv1.RegisterPermissionServiceHandler(ctx, mux, conn))
	must(pbresourcev1.RegisterResourceServiceHandler(ctx, mux, conn))
	must(pbgrouupv1.RegisterGroupServiceHandler(ctx, mux, conn))
	return mux
}

func must(err error) {
	if err != nil {
		logger.Severe(err)
	}
}
