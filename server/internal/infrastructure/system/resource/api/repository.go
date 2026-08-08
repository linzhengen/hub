package api

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/resource/api"
	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// New returns the API definitions the RBAC importer registers, read from the
// compiled protobuf descriptors.
//
// This used to parse a YAML file that `make gen` produced by reading the .proto
// files as text. The descriptors carry the same services and methods plus the
// hub.annotations.v1.method_rule option, which is how the importer learns that
// an rpc is public and must not get a permission record, and what to write as
// that permission's description. Reading them here also removes the second
// source of truth: the authorization interceptor already enforces off this
// catalog, so importer and enforcer can no longer disagree.
func New(catalog *apicatalog.Catalog) api.Repository {
	return &repositoryImpl{catalog: catalog}
}

type repositoryImpl struct {
	catalog *apicatalog.Catalog
}

func (r repositoryImpl) FindAll(_ context.Context) (api.APIs, error) {
	operations := r.catalog.Operations()
	apis := make(api.APIs, 0, len(operations))
	for _, op := range operations {
		apis = append(apis, &api.API{
			Service:  op.Service,
			Method:   op.Method,
			Resource: op.Resource,
			Action:   op.Action,
			Public:   op.Public,
			Summary:  op.Summary,
		})
	}
	return apis, nil
}
