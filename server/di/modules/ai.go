package modules

import (
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user"
	agentInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/agent"
	chatInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/chat"
	claudeInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/claude"
	mockInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/mock"
	toolInfra "github.com/linzhengen/hub/server/internal/infrastructure/ai/tool"
	aiHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler/ai"
	"github.com/linzhengen/hub/server/internal/interface/grpc/interceptor"
	aiUseCase "github.com/linzhengen/hub/server/internal/usecase/ai"
	pbaccessv1 "github.com/linzhengen/hub/server/pb/system/access/v1"
	pbauditv1 "github.com/linzhengen/hub/server/pb/system/audit/v1"
	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	pbpermissionv1 "github.com/linzhengen/hub/server/pb/system/permission/v1"
	pbresourcev1 "github.com/linzhengen/hub/server/pb/system/resource/v1"
	pbrolev1 "github.com/linzhengen/hub/server/pb/system/role/v1"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
	"github.com/linzhengen/hub/server/pkg/logger"
)

// ProvideAI registers AI feature dependencies.
func ProvideAI(c *dig.Container) {
	// infrastructure
	must(c.Provide(agentInfra.New))
	must(c.Provide(chatInfra.New))
	must(c.Provide(provideToolBox))
	must(c.Provide(func(cfg config.EnvConfig, tools chat.ToolBox) chat.Service {
		if cfg.Mock {
			logger.Infof("AI chat is using the mock LLM backend (ANTHROPIC_MOCK=true); no request reaches Anthropic")
			return mockInfra.New(cfg.Model, cfg.MockDelay, tools)
		}
		return claudeInfra.New(cfg.APIKey, cfg.Model, tools)
	}))
	// usecase
	must(c.Provide(aiUseCase.NewAgentUseCase))
	must(c.Provide(aiUseCase.NewChatUseCase))
	// interface (gRPC)
	must(c.Provide(aiHandler.NewAgentHandler))
	must(c.Provide(aiHandler.NewChatHandler))
}

// provideToolBox gives the assistant the read side of the API.
//
// The services are the same implementations the gRPC server registers, and the
// interceptors are the same authorization and validation ones it installs, so a
// tool call is checked exactly as the equivalent request would be rather than
// against a second copy of the rules. ChatService is deliberately not among
// them: pointing the assistant at its own sessions adds nothing and invites it
// to recurse.
func provideToolBox(
	catalog *apicatalog.Catalog,
	authSvc auth.Service,
	userRepo user.Repository,
	validator interceptor.Validator,
	auditRepo audit.Repository,
	transRepo trans.Repository,
	roleServiceServer pbrolev1.RoleServiceServer,
	userServiceServer pbuserv1.UserServiceServer,
	permissionServiceServer pbpermissionv1.PermissionServiceServer,
	resourceServiceServer pbresourcev1.ResourceServiceServer,
	groupServiceServer pbgroupv1.GroupServiceServer,
	auditServiceServer pbauditv1.AuditServiceServer,
	accessServiceServer pbaccessv1.AccessServiceServer,
	accessRequestServiceServer pbaccessv1.AccessRequestServiceServer,
) (chat.ToolBox, error) {
	return toolInfra.New(
		catalog,
		authSvc,
		[]toolInfra.Service{
			toolInfra.Register(&pbuserv1.UserService_ServiceDesc, userServiceServer),
			toolInfra.Register(&pbgroupv1.GroupService_ServiceDesc, groupServiceServer),
			toolInfra.Register(&pbrolev1.RoleService_ServiceDesc, roleServiceServer),
			toolInfra.Register(&pbpermissionv1.PermissionService_ServiceDesc, permissionServiceServer),
			toolInfra.Register(&pbresourcev1.ResourceService_ServiceDesc, resourceServiceServer),
			toolInfra.Register(&pbauditv1.AuditService_ServiceDesc, auditServiceServer),
			toolInfra.Register(&pbaccessv1.AccessService_ServiceDesc, accessServiceServer),
			toolInfra.Register(&pbaccessv1.AccessRequestService_ServiceDesc, accessRequestServiceServer),
		},
		interceptor.UnaryAuditInterceptor(auditRepo, transRepo, catalog),
		interceptor.UnaryAuthzInterceptor(authSvc, userRepo, catalog),
		interceptor.UnaryValidateInterceptor(validator),
	)
}
