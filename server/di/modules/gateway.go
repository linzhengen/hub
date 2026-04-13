package modules

import (
	"go.uber.org/dig"

	gwRegister "github.com/linzhengen/hub/server/internal/interface/grpcgw/register"
)


// ProvideGateway registers gRPC Gateway dependencies.
func ProvideGateway(c *dig.Container) {
	// interface (gRPC Gateway)
	must(c.Provide(gwRegister.New))
}
