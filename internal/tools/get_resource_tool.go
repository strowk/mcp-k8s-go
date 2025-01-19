package tools

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/k8s/apps/v1/deployment"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

func NewGetResourceTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"
	kindProperty := "kind"
	groupProperty := "group"
	versionProperty := "version"
	nameProperty := "name"
	templateProperty := "go_template"

	inputSchema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithRequiredString(namespaceProperty, "Namespace to get resource from"),
		toolinput.WithString(groupProperty, "API Group of the resource to get"),
		toolinput.WithString(versionProperty, "API Version of the resource to get"),
		toolinput.WithRequiredString(kindProperty, "Kind of resource to get"),
		toolinput.WithRequiredString(nameProperty, "Name of the resource to get"),
		toolinput.WithString(templateProperty, "Go template to render the output, if not specified, the complete JSON object will be returned"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "get-k8s-resource",
			Description: utils.Ptr("Get Kubernetes resource completely"),
			InputSchema: inputSchema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := inputSchema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace, err := input.String(namespaceProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			kind, err := input.String(kindProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			name, err := input.String(nameProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			group := input.StringOr(groupProperty, "")
			version := input.StringOr(versionProperty, "")

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			templateStr := input.StringOr(templateProperty, "")

			if strings.ToLower(kind) == "deployment" && (group == "apps" || group == "") && (version == "v1" || version == "") {
				return deployment.GetDeployment(clientset, namespace, name, templateStr)
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

						unstructured, err := dynClient.Resource(mapping.Resource).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
						if err != nil {
							return utils.ErrResponse(err)
						}

						object := unstructured.Object

						if metadata, ok := object["metadata"]; ok {
							if metadataMap, ok := metadata.(map[string]interface{}); ok {
								// this is too big and somewhat useless
								delete(metadataMap, "managedFields")
							}
						}

						var cnt any
						if templateStr != "" {
							tmpl, err := template.New("template").Parse(templateStr)
							if err != nil {
								return utils.ErrResponse(err)
							}
							buf := new(strings.Builder)
							err = tmpl.Execute(buf, object)
							if err != nil {
								return utils.ErrResponse(err)
							}
							cnt = mcp.TextContent{
								Type: "text",
								Text: buf.String(),
							}
						} else {
							c, err := content.NewJsonContent(object)
							if err != nil {
								return utils.ErrResponse(err)
							}
							cnt = c
						}
						var contents = []any{cnt}

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
