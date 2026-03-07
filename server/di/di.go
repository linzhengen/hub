package di

import (
	"database/sql"
	"log"

	"github.com/Nerzal/gocloak/v13"
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	systemDomain "github.com/linzhengen/hub/server/internal/domain/system"
	"github.com/linzhengen/hub/server/internal/domain/user"
	authInfra "github.com/linzhengen/hub/server/internal/infrastructure/auth"
	oidcAdminInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
	tokenInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/token"
	oidcUserInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
	groupInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/group"
	grouproleInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/group/grouprole"
	permissionInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/permission"
	resourceInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource"
	apiInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource/api"
	menuInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource/menu"
	roleInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/role"
	rolepermissionInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/role/rolepermission"
	transInfra "github.com/linzhengen/hub/server/internal/infrastructure/trans"
	userInfra "github.com/linzhengen/hub/server/internal/infrastructure/user"
	usergroupInfra "github.com/linzhengen/hub/server/internal/infrastructure/user/usergroup"
	cmdHandler "github.com/linzhengen/hub/server/internal/interface/cmd/handler"
	cmdRegister "github.com/linzhengen/hub/server/internal/interface/cmd/register"
	grpcHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler"
	systemHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler/system"
	"github.com/linzhengen/hub/server/internal/interface/grpc/register"
	gwRegister "github.com/linzhengen/hub/server/internal/interface/grpcgw/register"
	"github.com/linzhengen/hub/server/internal/usecase"
	"github.com/linzhengen/hub/server/internal/usecase/develop"
	"github.com/linzhengen/hub/server/internal/usecase/system"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func NewDI(envCfg config.EnvConfig, db *sql.DB) *dig.Container {
	c := dig.New()
	// config
	must(c.Provide(func() config.MySQL {
		return envCfg.MySQL
	}))
	must(c.Provide(func() config.EnvConfig {
		return envCfg
	}))
	must(c.Provide(func() config.KeyCloak {
		return envCfg.KeyCloak
	}))
	// grpc server options
	must(c.Provide(func() *register.Opts {
		return &register.Opts{
			APIRateLimit:       envCfg.ApiRateLimit,
			MaxGRPCMessageSize: envCfg.MaxGRPCMessageSize,
			DisableAuth:        envCfg.DisableAuth,
		}
	}))
	// db
	must(c.Provide(func() *sql.DB {
		return db
	}))
	must(c.Provide(func() sqlc.DBTX {
		return db
	}))
	must(c.Provide(mysql.NewDialect))

	// domain
	must(c.Provide(user.NewService))
	must(c.Provide(auth.NewService))
	must(c.Provide(systemDomain.NewResourceService))

	// infrastructure
	must(c.Provide(groupInfra.New))
	must(c.Provide(permissionInfra.New))
	must(c.Provide(resourceInfra.New))
	must(c.Provide(roleInfra.New))
	must(c.Provide(userInfra.New))
	must(c.Provide(userInfra.NewFinder))
	must(c.Provide(authInfra.NewRepository))
	must(c.Provide(sqlc.New))
	must(c.Provide(transInfra.New))
	must(c.Provide(grouproleInfra.New))
	must(c.Provide(rolepermissionInfra.New))
	must(c.Provide(usergroupInfra.New))
	must(c.Provide(oidcAdminInfra.NewClient))
	must(c.Provide(oidcUserInfra.NewRepository))
	must(c.Provide(func() *tokenInfra.KeyCloak {
		return &tokenInfra.KeyCloak{
			ClientId:    envCfg.ClientId,
			Realm:       envCfg.Realm,
			Client:      gocloak.NewClient(envCfg.KeycloakURL),
			DisableAuth: envCfg.DisableAuth,
		}
	}))
	must(c.Provide(tokenInfra.New))
	must(c.Provide(apiInfra.New))
	must(c.Provide(menuInfra.New))

	// usecase
	must(c.Provide(system.NewGroupUseCase))
	must(c.Provide(system.NewPermissionUseCase))
	must(c.Provide(system.NewResourceUseCase))
	must(c.Provide(system.NewRoleUseCase))
	must(c.Provide(develop.NewResourceUseCase))
	must(c.Provide(usecase.NewUserUseCase))
	must(c.Provide(develop.NewMigrateUseCase))
	must(c.Provide(develop.NewSeedUseCase))

	// interface (gRPC)
	must(c.Provide(systemHandler.NewGroupHandler))
	must(c.Provide(systemHandler.NewPermissionHandler))
	must(c.Provide(systemHandler.NewResourceHandler))
	must(c.Provide(systemHandler.NewRoleHandler))
	must(c.Provide(grpcHandler.NewUserHandler))
	must(c.Provide(register.New))

	// interface (gRPC Gateway)
	must(c.Provide(gwRegister.New))

	// interface (cmd)
	must(c.Provide(cmdHandler.NewMigrateHandler))
	must(c.Provide(cmdHandler.NewSeedHandler))
	must(c.Provide(cmdHandler.NewImportResourceHandler))
	must(c.Provide(cmdRegister.New))
	return c
}
