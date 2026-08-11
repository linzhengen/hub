package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	chatInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/chat"
	claudeInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/claude"
	chatHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler/ai"
	aiUseCase "github.com/linzhengen/hub/server/internal/usecase/ai"
)

// ProvideAI registers AI feature dependencies.
func ProvideAI(c *dig.Container) {
	// infrastructure
	must(c.Provide(chatInfra.New))
	must(c.Provide(func(cfg config.EnvConfig) chat.Service {
		return claudeInfra.New(cfg.APIKey, cfg.Model)
	}))
	// usecase
	must(c.Provide(aiUseCase.NewChatUseCase))
	// interface (gRPC)
	must(c.Provide(chatHandler.NewChatHandler))
}
