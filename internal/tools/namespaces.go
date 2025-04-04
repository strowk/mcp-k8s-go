package tools

import (
	"context"
	"sort"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListNamespacesTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-namespaces",
			Description: utils.Ptr("List Kubernetes namespaces using specific context"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}
			k8sCtx := input.StringOr(contextProperty, "")

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return errResponse(err)
			}

			namespace, err := clientset.
				CoreV1().
				Namespaces().
				List(ctx, metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

			sort.Slice(namespace.Items, func(i, j int) bool {
				return namespace.Items[i].Name < namespace.Items[j].Name
			})

			var contents = make([]interface{}, len(namespace.Items))
			for i, namespace := range namespace.Items {
				content, err := NewJsonContent(NamespacesInList{
					Name: namespace.Name,
				})
				if err != nil {
					return errResponse(err)
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

type NamespacesInList struct {
	Name string `json:"name"`
}
