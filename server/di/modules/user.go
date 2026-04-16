package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	userInfra "github.com/linzhengen/hub/v1/server/internal/infrastructure/user"
	usergroupInfra "github.com/linzhengen/hub/v1/server/internal/infrastructure/user/usergroup"
	"github.com/linzhengen/hub/v1/server/internal/interface/grpc/handler"
	"github.com/linzhengen/hub/v1/server/internal/usecase"
)

// ProvideUser registers user-related dependencies.
func ProvideUser(c *dig.Container) {
	// domain
	must(c.Provide(user.NewService))
	// infrastructure
	must(c.Provide(userInfra.New))
	must(c.Provide(userInfra.NewFinder))
	must(c.Provide(usergroupInfra.New))
	// usecase
	must(c.Provide(usecase.NewUserUseCase))
	// interface (gRPC)
	must(c.Provide(handler.NewUserHandler))
}
