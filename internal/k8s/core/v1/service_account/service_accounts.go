package serviceaccount

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

func NewListServiceAccountsTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list Service Accounts from, defaults to all namespaces"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-service-accounts",
			Description: utils.Ptr("List Kubernetes Service Accounts with detailed information"),
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

			var serviceAccounts *v1.ServiceAccountList
			if namespace == "" {
				// List Service Accounts from all namespaces
				serviceAccounts, err = clientset.CoreV1().ServiceAccounts(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			} else {
				// List Service Accounts from specific namespace
				serviceAccounts, err = clientset.CoreV1().ServiceAccounts(namespace).List(context.Background(), metav1.ListOptions{})
			}

			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(serviceAccounts.Items, func(i, j int) bool {
				// Sort by namespace, then by name
				if serviceAccounts.Items[i].Namespace == serviceAccounts.Items[j].Namespace {
					return serviceAccounts.Items[i].Name < serviceAccounts.Items[j].Name
				}
				return serviceAccounts.Items[i].Namespace < serviceAccounts.Items[j].Namespace
			})

			var contents []interface{} = make([]interface{}, len(serviceAccounts.Items))
			for i, sa := range serviceAccounts.Items {
				// Calculate age
				age := time.Since(sa.CreationTimestamp.Time)

				// Count secrets and image pull secrets
				secretsCount := len(sa.Secrets)
				imagePullSecretsCount := len(sa.ImagePullSecrets)

				content, err := content.NewJsonContent(ServiceAccountDetails{
					Name:             sa.Name,
					Namespace:        sa.Namespace,
					Secrets:          secretsCount,
					ImagePullSecrets: imagePullSecretsCount,
					Age:              utils.FormatAge(age),
					CreatedAt:        sa.CreationTimestamp.Format(time.RFC3339),
					AutomountToken:   getAutomountTokenStatus(&sa),
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

// ServiceAccountDetails provides a structured representation of Service Account information
type ServiceAccountDetails struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	Secrets          int    `json:"secrets"`
	ImagePullSecrets int    `json:"imagePullSecrets"`
	Age              string `json:"age"`
	CreatedAt        string `json:"createdAt"`
	AutomountToken   string `json:"automountToken"`
}

// getAutomountTokenStatus determines the automount token status
func getAutomountTokenStatus(sa *v1.ServiceAccount) string {
	if sa.AutomountServiceAccountToken == nil {
		return "Default"
	}

	if *sa.AutomountServiceAccountToken {
		return "Enabled"
	}

	return "Disabled"
}
