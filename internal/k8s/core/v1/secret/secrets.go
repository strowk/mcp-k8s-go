package secret

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

func NewListSecretsTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list Secrets from, defaults to all namespaces"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-secrets",
			Description: utils.Ptr("List Kubernetes Secrets with detailed information"),
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

			var secrets *v1.SecretList
			if namespace == "" {
				// List Secrets from all namespaces
				secrets, err = clientset.CoreV1().Secrets(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			} else {
				// List Secrets from specific namespace
				secrets, err = clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
			}

			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(secrets.Items, func(i, j int) bool {
				// Sort by namespace, then by name
				if secrets.Items[i].Namespace == secrets.Items[j].Namespace {
					return secrets.Items[i].Name < secrets.Items[j].Name
				}
				return secrets.Items[i].Namespace < secrets.Items[j].Namespace
			})

			var contents []interface{} = make([]interface{}, len(secrets.Items))
			for i, secret := range secrets.Items {
				// Calculate age
				age := time.Since(secret.CreationTimestamp.Time)

				content, err := content.NewJsonContent(SecretInList{
					Name:      secret.Name,
					Namespace: secret.Namespace,
					Age:       utils.FormatAge(age),
					Type:      string(secret.Type),
					KeysCount: len(secret.Data),
					CreatedAt: secret.CreationTimestamp.Format(time.RFC3339),
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

// SecretInList provides a structured representation of Secret information
type SecretInList struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Type      string `json:"type"`
	KeysCount int    `json:"keys_count"`
	CreatedAt string `json:"createdAt"`
}
