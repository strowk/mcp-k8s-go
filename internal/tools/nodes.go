package tools

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListNodesTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-nodes",
			Description: utils.Ptr("List Kubernetes nodes using specific context"),
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

			nodes, err := clientset.
				CoreV1().
				Nodes().
				List(ctx, metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

			sort.Slice(nodes.Items, func(i, j int) bool {
				return nodes.Items[i].Name < nodes.Items[j].Name
			})

			var contents = make([]interface{}, len(nodes.Items))
			for i, ns := range nodes.Items {
				// Calculate age
				age := time.Since(ns.CreationTimestamp.Time)

				// Determine status
				status := "NotReady"
				for _, condition := range ns.Status.Conditions {
					if condition.Type == "Ready" {
						if condition.Status == "True" {
							status = "Ready"
						} else {
							status = "NotReady"
						}
						break
					}
				}

				content, err := NewJsonContent(NodeInList{
					Name:      ns.Name,
					Status:    status,
					Age:       formatAge(age),
					CreatedAt: ns.CreationTimestamp.Time,
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

// NodeInList provides a structured representation of node information
type NodeInList struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Age       string    `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// formatAge converts a duration to a human-readable age string
func formatAge(duration time.Duration) string {
	if duration.Hours() < 1 {
		return duration.Round(time.Minute).String()
	}
	if duration.Hours() < 24 {
		return duration.Round(time.Hour).String()
	}
	days := int(duration.Hours() / 24)
	return formatDays(days)
}

// formatDays provides a concise representation of days
func formatDays(days int) string {
	if days < 7 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 30 {
		weeks := days / 7
		return fmt.Sprintf("%dw", weeks)
	}
	months := days / 30
	return fmt.Sprintf("%dmo", months)
}
