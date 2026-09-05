package modules

import (
	"go.uber.org/dig"

	systemDomain "github.com/linzhengen/hub/server/internal/domain/system"
	accessInfra "github.com/linzhengen/hub/server/internal/infrastructure/access"
	auditInfra "github.com/linzhengen/hub/server/internal/infrastructure/audit"
	delegationInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/delegation"
	groupInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/group"
	grouproleInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/group/grouprole"
	organizationInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/organization"
	permissionInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/permission"
	resourceInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource"
	apiInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource/api"
	menuInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/resource/menu"
	roleInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/role"
	rolepermissionInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/role/rolepermission"
	serviceAccountInfra "github.com/linzhengen/hub/server/internal/infrastructure/system/serviceaccount"
	systemHandler "github.com/linzhengen/hub/server/internal/interface/grpc/handler/system"
	"github.com/linzhengen/hub/server/internal/usecase/develop"
	"github.com/linzhengen/hub/server/internal/usecase/system"
)

// ProvideSystem registers system-related dependencies.
func ProvideSystem(c *dig.Container) {
	// domain
	must(c.Provide(systemDomain.NewResourceService))
	// infrastructure
	must(c.Provide(accessInfra.New))
	must(c.Provide(auditInfra.New))
	must(c.Provide(delegationInfra.New))
	must(c.Provide(groupInfra.New))
	must(c.Provide(organizationInfra.New))
	must(c.Provide(permissionInfra.New))
	must(c.Provide(resourceInfra.New))
	must(c.Provide(roleInfra.New))
	must(c.Provide(serviceAccountInfra.New))
	must(c.Provide(grouproleInfra.New))
	must(c.Provide(rolepermissionInfra.New))
	must(c.Provide(apiInfra.New))
	must(c.Provide(menuInfra.New))
	// usecase
	must(c.Provide(system.NewAccessUseCase))
	must(c.Provide(system.NewAccessRequestUseCase))
	must(c.Provide(system.NewAuditUseCase))
	must(c.Provide(system.NewDelegationUseCase))
	must(c.Provide(system.NewGroupUseCase))
	must(c.Provide(system.NewOrganizationUseCase))
	must(c.Provide(system.NewPermissionUseCase))
	must(c.Provide(system.NewResourceUseCase))
	must(c.Provide(system.NewRoleUseCase))
	must(c.Provide(system.NewServiceAccountUseCase))
	must(c.Provide(develop.NewResourceUseCase))
	// interface (gRPC)
	must(c.Provide(systemHandler.NewAccessHandler))
	must(c.Provide(systemHandler.NewAccessRequestHandler))
	must(c.Provide(systemHandler.NewAuditHandler))
	must(c.Provide(systemHandler.NewDelegationHandler))
	must(c.Provide(systemHandler.NewGroupHandler))
	must(c.Provide(systemHandler.NewOrganizationHandler))
	must(c.Provide(systemHandler.NewPermissionHandler))
	must(c.Provide(systemHandler.NewResourceHandler))
	must(c.Provide(systemHandler.NewRoleHandler))
	must(c.Provide(systemHandler.NewServiceAccountHandler))
}
