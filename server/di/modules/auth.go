package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	authInfra "github.com/linzhengen/hub/server/internal/infrastructure/auth"
	oidcAdminInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
	oidcUserInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/interface/grpc/interceptor"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// ProvideAuth registers authentication and authorization dependencies.
func ProvideAuth(c *dig.Container) {
	// domain
	must(c.Provide(auth.NewService))
	// the RBAC rules the authz interceptor enforces, read off the protos
	must(c.Provide(apicatalog.Default))
	// request constraints, likewise read off the protos
	must(c.Provide(interceptor.NewValidator))
	// infrastructure
	must(c.Provide(func(q persistence.Querier, envCfg config.EnvConfig) auth.Repository {
		return authInfra.NewCachingRepository(authInfra.NewRepository(q), envCfg.AuthzPolicyCacheTTL)
	}))
	must(c.Provide(oidcAdminInfra.NewClient))
	must(c.Provide(oidcUserInfra.NewRepository))
}
