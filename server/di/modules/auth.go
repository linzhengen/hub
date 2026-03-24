package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	authInfra "github.com/linzhengen/hub/server/internal/infrastructure/auth"
	oidcAdminInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
	oidcUserInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/user"
)


// ProvideAuth registers authentication and authorization dependencies.
func ProvideAuth(c *dig.Container) {
	// domain
	must(c.Provide(auth.NewService))
	// infrastructure
	must(c.Provide(authInfra.NewRepository))
	must(c.Provide(oidcAdminInfra.NewClient))
	must(c.Provide(oidcUserInfra.NewRepository))
}
