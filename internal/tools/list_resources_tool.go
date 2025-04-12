package tools

import (
	"context"
	"sort"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewListResourcesTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"
	kindProperty := "kind"
	groupProperty := "group"
	versionProperty := "version"

	inputSchema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list resources from, defaults to all namespaces"),
		toolinput.WithString(groupProperty, "API Group of resources to list"),
		toolinput.WithString(versionProperty, "API Version of resources to list"),
		toolinput.WithRequiredString(kindProperty, "Kind of resources to list"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-resources",
			Description: utils.Ptr("List arbitrary Kubernetes resources"),
			InputSchema: inputSchema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
			input, err := inputSchema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace := input.StringOr(namespaceProperty, metav1.NamespaceAll)

			kind, err := input.String(kindProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			group := input.StringOr(groupProperty, "")
			version := input.StringOr(versionProperty, "")

			informer, err := pool.GetInformer(k8sCtx, kind, group, version)
			if err != nil {
				return utils.ErrResponse(err)
			}

			listMapping := pool.GetListMapping(k8sCtx, kind, group, version)
			var unstructuredList []runtime.Object

			if namespace != metav1.NamespaceAll {
				unstructured, err := informer.Lister().ByNamespace(namespace).List(labels.Everything())
				if err != nil {
					return utils.ErrResponse(err)
				}
				unstructuredList = unstructured
			} else {
				unstructured, err := informer.Lister().List(labels.Everything())
				if err != nil {
					return utils.ErrResponse(err)
				}
				unstructuredList = unstructured
			}

			var contents = make([]any, 0)
			var listContents []list_mapping.ListContentItem
			for _, unstructuredItem := range unstructuredList {
				item := unstructuredItem.(runtime.Unstructured)
				var listContent list_mapping.ListContentItem

				if listMapping == nil {
					unscructuredContent := item.UnstructuredContent()

					// we try to list only metadata to avoid too big outputs
					metadata, ok := unscructuredContent["metadata"].(map[string]any)
					if !ok {
						cnt, err := content.NewJsonContent(unscructuredContent)
						if err != nil {
							return utils.ErrResponse(err)
						}
						contents = append(contents, cnt)
						continue
					}
					gen := GenericListContent{}
					if name, ok := metadata["name"].(string); ok {
						gen.Name = name
					}

					if namespace, ok := metadata["namespace"].(string); ok {
						gen.Namespace = namespace
					}

					listContent = gen
				} else {
					listContent, err = listMapping(item)
					if err != nil {
						return utils.ErrResponse(err)
					}
				}
				listContents = append(listContents, listContent)
			}

			// sort the list contents by name and namespace
			sort.Slice(listContents, func(i, j int) bool {
				if listContents[i].GetNamespace() == listContents[j].GetNamespace() {
					return listContents[i].GetName() < listContents[j].GetName()
				}
				return listContents[i].GetNamespace() < listContents[j].GetNamespace()
			})

			// convert sorted list contents to JSON content
			for _, listContent := range listContents {
				cnt, err := content.NewJsonContent(listContent)
				if err != nil {
					return utils.ErrResponse(err)
				}
				contents = append(contents, cnt)
			}

			return &mcp.CallToolResult{
				Meta:    map[string]any{},
				Content: contents,
				IsError: utils.Ptr(false),
			}
		},
	)
}

type GenericListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

func (l GenericListContent) GetName() string {
	return l.Name
}

func (l GenericListContent) GetNamespace() string {
	return l.Namespace
}
