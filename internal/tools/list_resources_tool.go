package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/k8s/apps/v1/deployment"
	"github.com/strowk/mcp-k8s-go/internal/k8s/core/v1/pod"
	"github.com/strowk/mcp-k8s-go/internal/k8s/core/v1/service"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
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

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			if strings.ToLower(kind) == "deployment" && (group == "apps" || group == "") && (version == "v1" || version == "") {
				return deployment.ListDeployments(clientset, namespace)
			}

			if strings.ToLower(kind) == "pod" && (group == "" || group == "core") && (version == "" || version == "v1") {
				return pod.ListPods(clientset, namespace)
			}

			if strings.ToLower(kind) == "service" && (group == "" || group == "core") && (version == "" || version == "v1") {
				return service.ListServices(clientset, namespace)
			}

			res, err := clientset.Discovery().ServerPreferredResources()
			if res == nil && err != nil {
				return utils.ErrResponse(err)
			}

			cfg := k8s.GetKubeConfigForContext(k8sCtx)
			restConfig, err := cfg.ClientConfig()
			if err != nil {
				return utils.ErrResponse(err)
			}

			dynClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return utils.ErrResponse(err)
			}
			mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))

			for _, r := range res {
				for _, apiResource := range r.APIResources {
					if strings.EqualFold(apiResource.Kind, kind) && (group == "" || strings.EqualFold(apiResource.Group, group)) && (version == "" || strings.EqualFold(apiResource.Version, version)) {
						gvk := schema.GroupVersionKind{
							Group:   apiResource.Group,
							Version: apiResource.Version,
							Kind:    apiResource.Kind,
						}

						mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
						if err != nil {
							return utils.ErrResponse(err)
						}

						unstructured, err := dynClient.Resource(mapping.Resource).Namespace(namespace).List(ctx, metav1.ListOptions{})
						if err != nil {
							return utils.ErrResponse(err)
						}

						var contents = make([]interface{}, len(unstructured.Items))
						for i, item := range unstructured.Items {
							// we try to list only metadata to avoid too big outputs
							metadata, ok := item.Object["metadata"].(map[string]interface{})
							if !ok {
								cnt, err := content.NewJsonContent(metadata)
								if err != nil {
									return utils.ErrResponse(err)
								}
								contents[i] = cnt
								continue
							}

							listContent := ListContent{}
							if name, ok := metadata["name"].(string); ok {
								listContent.Name = name
							}

							if namespace, ok := metadata["namespace"].(string); ok {
								listContent.Namespace = namespace
							}

							cnt, err := content.NewJsonContent(listContent)
							if err != nil {
								return utils.ErrResponse(err)
							}
							contents[i] = cnt
						}

						return &mcp.CallToolResult{
							Meta:    map[string]interface{}{},
							Content: contents,
							IsError: utils.Ptr(false),
						}
					}
				}
			}

			return utils.ErrResponse(fmt.Errorf("resource %s/%s/%s not found", group, version, kind))
		},
	)
}

type ListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
