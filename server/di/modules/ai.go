package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	chatInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/chat"
	claudeInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/claude"
	mockInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/mock"
	chatHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler/ai"
	aiUseCase "github.com/linzhengen/hub/server/internal/usecase/ai"
	"github.com/linzhengen/hub/server/pkg/logger"
)

// ProvideAI registers AI feature dependencies.
func ProvideAI(c *dig.Container) {
	// infrastructure
	must(c.Provide(chatInfra.New))
	must(c.Provide(func(cfg config.EnvConfig) chat.Service {
		if cfg.Anthropic.Mock {
			logger.Infof("AI chat is using the mock LLM backend (ANTHROPIC_MOCK=true); no request reaches Anthropic")
			return mockInfra.New(cfg.Anthropic.Model, cfg.Anthropic.MockDelay)
		}
		return claudeInfra.New(cfg.Anthropic.APIKey, cfg.Anthropic.Model)
	}))
	// usecase
	must(c.Provide(aiUseCase.NewChatUseCase))
	// interface (gRPC)
	must(c.Provide(chatHandler.NewChatHandler))
}
