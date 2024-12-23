package configmap

import (
	"context"
	"sort"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListConfigMapsTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list ConfigMaps from, defaults to all namespaces"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-configmaps",
			Description: utils.Ptr("List Kubernetes ConfigMaps with detailed information"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace := input.StringOr(namespaceProperty, "")

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			var configMaps *v1.ConfigMapList
			if namespace == "" {
				// List ConfigMaps from all namespaces
				configMaps, err = clientset.CoreV1().ConfigMaps(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			} else {
				// List ConfigMaps from specific namespace
				configMaps, err = clientset.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
			}

			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(configMaps.Items, func(i, j int) bool {
				// Sort by namespace, then by name
				if configMaps.Items[i].Namespace == configMaps.Items[j].Namespace {
					return configMaps.Items[i].Name < configMaps.Items[j].Name
				}
				return configMaps.Items[i].Namespace < configMaps.Items[j].Namespace
			})

			var contents []interface{} = make([]interface{}, len(configMaps.Items))
			for i, cm := range configMaps.Items {
				// Calculate age
				age := time.Since(cm.CreationTimestamp.Time)

				content, err := content.NewJsonContent(ConfigMapInList{
					Name:      cm.Name,
					Namespace: cm.Namespace,
					Age:       utils.FormatAge(age),
					KeysCount: len(cm.Data),
					CreatedAt: cm.CreationTimestamp.Format(time.RFC3339),
				})
				if err != nil {
					return utils.ErrResponse(err)
				}
				contents[i] = content
			}

			return &mcp.CallToolResult{
				Meta:    map[string]interface{}{},
				Content: contents,
				IsError: utils.Ptr(false),
			}
		},
	)
}

// ConfigMapInList provides a structured representation of ConfigMap information
type ConfigMapInList struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	KeysCount int    `json:"keys_count"`
	CreatedAt string `json:"createdAt"`
}
