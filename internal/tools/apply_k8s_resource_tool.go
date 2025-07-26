package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

func NewApplyK8sResourceTool(clientPool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	manifestProperty := "manifest"

	inputSchema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithRequiredString(manifestProperty, "YAML manifest of the resource to apply"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "apply-k8s-resource",
			Description: utils.Ptr("Create or modify a Kubernetes resource from a YAML manifest"),
			InputSchema: inputSchema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]any) *mcp.CallToolResult {
			input, err := inputSchema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			clientset, err := clientPool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			manifest, err := input.String(manifestProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			dynamicClient, err := clientPool.GetDynamicClient(k8sCtx)
			if err != nil {
				return utils.ErrResponse(fmt.Errorf("failed to retrieve dynamic client from the client pool: %w", err))
			}

			decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manifest)), 4096)
			var results []string
			for {
				obj := &unstructured.Unstructured{}
				if err := decoder.Decode(obj); err != nil {
					if err == io.EOF {
						break
					}
					return utils.ErrResponse(fmt.Errorf("failed to unmarshal manifest: %w", err))
				}

				if obj.Object == nil {
					continue
				}

				namespace := obj.GetNamespace()
				if namespace == "" {
					namespace = "default"
				}

				gvk := obj.GroupVersionKind()
				apiResource, err := findAPIResource(clientset.Discovery(), gvk)
				if err != nil {
					return utils.ErrResponse(fmt.Errorf("failed to find API resource: %w", err))
				}

				var dr dynamic.ResourceInterface
				if apiResource.Namespaced {
					dr = dynamicClient.Resource(gvk.GroupVersion().WithResource(apiResource.Name)).Namespace(namespace)
				} else {
					dr = dynamicClient.Resource(gvk.GroupVersion().WithResource(apiResource.Name))
				}
				action := "configured"

				if _, err := dr.Get(ctx, obj.GetName(), metav1.GetOptions{}); errors.IsNotFound(err) {
					action = "created"
				}

				rawObj, err := obj.MarshalJSON()
				if err != nil {
					return utils.ErrResponse(fmt.Errorf("failed to marshal resource: %w", err))
				}

				_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, rawObj, metav1.PatchOptions{
					FieldManager: "mcp-k8s-go",
				})

				if err != nil {
					return utils.ErrResponse(fmt.Errorf("failed to patch resource: %w", err))
				}

				resourceText := fmt.Sprintf("%s.%s/%s", strings.ToLower(apiResource.Name), gvk.Group, obj.GetName())
				if gvk.Group == "" {
					resourceText = fmt.Sprintf("%s/%s", strings.ToLower(apiResource.Name), obj.GetName())
				}
				results = append(results, fmt.Sprintf("%s %s", resourceText, action))
			}

			return &mcp.CallToolResult{
				Content: []any{
					mcp.TextContent{
						Type: "text",
						Text: strings.Join(results, "\n"),
					},
				},
			}
		},
	)
}

func findAPIResource(discoveryClient discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	apiResourceList, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources for group version %s: %w", gvk.GroupVersion().String(), err)
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Kind == gvk.Kind {
			return &apiResource, nil
		}
	}

	return nil, fmt.Errorf("resource kind %s not found in group version %s", gvk.Kind, gvk.GroupVersion().String())
}
