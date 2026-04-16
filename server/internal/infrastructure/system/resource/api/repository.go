package api

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource/api"
	yamlData "github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/yaml/proto"
)

func New() api.Repository {
	return &repositoryImpl{}
}

type repositoryImpl struct {
}

func (r repositoryImpl) FindAll(_ context.Context) (api.APIs, error) {
	protoServices := yamlData.SelectAllServices()
	apis := make(api.APIs, 0)

	for _, service := range protoServices {
		for _, method := range service.Methods {
			apis = append(apis, &api.API{
				Service: service.Service,
				Method:  method.Name,
			})
		}
	}

	return apis, nil
}
