package tools

import (
	"context"
	"encoding/json"
	"fmt"
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

			obj := &unstructured.Unstructured{}
			if err := yaml.Unmarshal([]byte(manifest), obj); err != nil {
				return utils.ErrResponse(fmt.Errorf("failed to unmarshal manifest: %w", err))
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

			dynamicClient, err := clientPool.GetDynamicClient(k8sCtx)
			if err != nil {
				return utils.ErrResponse(fmt.Errorf("failed to retrieve dynamic client from the client pool: %w", err))
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

			patchBytes, err := json.Marshal(obj)
			if err != nil {
				return utils.ErrResponse(fmt.Errorf("failed to marshal manifest to JSON: %w", err))
			}
			_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, patchBytes, metav1.PatchOptions{
				FieldManager: "mcp-k8s-go",
			})

			if err != nil {
				return utils.ErrResponse(fmt.Errorf("failed to patch resource: %w", err))
			}

			resourceIdentifier := fmt.Sprintf("%s.%s/%s", strings.ToLower(apiResource.Name), gvk.Group, obj.GetName())
			if gvk.Group == "" {
				resourceIdentifier = fmt.Sprintf("%s/%s", strings.ToLower(apiResource.Name), obj.GetName())
			}

			return &mcp.CallToolResult{
				Content: []any{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("%s %s", resourceIdentifier, action),
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
