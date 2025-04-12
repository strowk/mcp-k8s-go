package list_mapping

import (
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ListMapping func(u runtime.Unstructured) (ListContentItem, error)

type ListContentItem interface {
	GetName() string
	GetNamespace() string
}

type ListMappingResolver interface {
	GetListMapping(gvk *schema.GroupVersionKind) ListMapping
}

const (
	MappingResolversTag = `group:"list_mapping_resolvers"`
)

func AsMappingResolver(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(ListMappingResolver)), fx.ResultTags(MappingResolversTag),
	)
}
