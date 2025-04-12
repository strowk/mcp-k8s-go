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
	"github.com/strowk/mcp-k8s-go/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		toolinput.WithString(namespaceProperty, "Namespace to get resource from, skip for cluster resources"),
		toolinput.WithString(groupProperty, "API Group of the resource to get"),
		toolinput.WithString(versionProperty, "API Version of the resource to get"),
		toolinput.WithRequiredString(kindProperty, "Kind of resource to get"),
		toolinput.WithRequiredString(nameProperty, "Name of the resource to get"),
		toolinput.WithString(templateProperty, "Go template to render the output, if not specified, the complete JSON object will be returned"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "get-k8s-resource",
			Description: utils.Ptr("Get details of any Kubernetes resource like pod, node or service - completely as JSON or rendered using template"),
			InputSchema: inputSchema.GetMcpToolInputSchema(),
		},
		func(_ context.Context, args map[string]any) *mcp.CallToolResult {
			input, err := inputSchema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace := input.StringOr(namespaceProperty, "")

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

			templateStr := input.StringOr(templateProperty, "")

			informer, err := pool.GetInformer(k8sCtx, kind, group, version)
			if err != nil {
				return utils.ErrResponse(err)
			}

			var key string

			if namespace == "" {
				key = name
			} else {
				key = fmt.Sprintf("%s/%s", namespace, name)
			}
			accumulator, exist, err := informer.Informer().GetIndexer().GetByKey(key)
			if err != nil {
				return utils.ErrResponse(err)
			}
			if !exist {
				return utils.ErrResponse(fmt.Errorf("resource %s/%s/%s/%s/%s not found", group, version, kind, namespace, name))
			}
			unstructuredAcc, ok := accumulator.(*unstructured.Unstructured)
			if !ok {
				return utils.ErrResponse(fmt.Errorf("resource %s/%s/%s/%s/%s is not unstructured", group, version, kind, namespace, name))
			}

			object := unstructuredAcc.Object

			if metadata, ok := object["metadata"]; ok {
				if metadataMap, ok := metadata.(map[string]any); ok {
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
				Meta:    map[string]any{},
				Content: contents,
				IsError: utils.Ptr(false),
			}
		},
	)
}
