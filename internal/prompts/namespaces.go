package prompts

import (
	"context"
	"fmt"
	"sort"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListNamespacesPrompt(pool k8s.ClientPool) fxctx.Prompt {
	return fxctx.NewPrompt(
		mcp.Prompt{
			Name: "list-k8s-namespaces",
			Description: utils.Ptr(
				"List Kubernetes Namespaces in the specified context",
			),
			Arguments: []mcp.PromptArgument{
				{
					Name: "context",
					Description: utils.Ptr(
						"Context to list namespaces in, defaults to current context",
					),
					Required: utils.Ptr(false),
				},
			},
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			k8sContext := req.Params.Arguments["context"]
			clientset, err := pool.GetClientset(k8sContext)
			if err != nil {
				return nil, fmt.Errorf("failed to get k8s client: %w", err)
			}

			namespaces, err := clientset.
				CoreV1().
				Namespaces().
				List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list namespaces: %w", err)
			}

			sort.Slice(namespaces.Items, func(i, j int) bool {
				return namespaces.Items[i].Name < namespaces.Items[j].Name
			})

			ofContextMsg := ""
			currentContext, err := k8s.GetCurrentContext()
			if err == nil && currentContext != "" {
				ofContextMsg = fmt.Sprintf(", context '%s'", currentContext)
			}

			var messages = make(
				[]mcp.PromptMessage,
				len(namespaces.Items)+1,
			)
			messages[0] = mcp.PromptMessage{
				Content: mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf(
						"There are %d namespaces%s:",
						len(namespaces.Items),
						ofContextMsg,
					),
				},
				Role: mcp.RoleUser,
			}

			type NamespaceInList struct {
				Name string `json:"name"`
			}

			for i, namespace := range namespaces.Items {
				content, err := content.NewJsonContent(NamespaceInList{
					Name: namespace.Name,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create content: %w", err)
				}
				messages[i+1] = mcp.PromptMessage{
					Content: content,
					Role:    mcp.RoleUser,
				}
			}

			return &mcp.GetPromptResult{
				Description: utils.Ptr(
					fmt.Sprintf("Namespaces%s", ofContextMsg),
				),
				Messages: messages,
			}, nil
		},
	)
}
