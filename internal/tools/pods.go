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

func NewListPodsTool(pool k8s.ClientPool) fxctx.Tool {
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString("context", "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString("namespace", "Namespace to list pods from, defaults to all namespaces"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-pods",
			Description: utils.Ptr("List Kubernetes pods using specific context in a specified namespace"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}

			k8sCtx := input.StringOr("context", "")
			k8sNamespace := input.StringOr("namespace", "")
			if k8sNamespace == "" {
				k8sNamespace = metav1.NamespaceAll
			}

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return errResponse(err)
			}

			pods, err := clientset.
				CoreV1().
				Pods(k8sNamespace).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

			sort.Slice(pods.Items, func(i, j int) bool {
				return pods.Items[i].Name < pods.Items[j].Name
			})

			var contents []interface{} = make([]interface{}, len(pods.Items))
			for i, pod := range pods.Items {
				content, err := NewJsonContent(PodInList{
					Name:      pod.Name,
					Namespace: pod.Namespace,
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

type PodInList struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func errResponse(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: utils.Ptr(true),
		Meta:    map[string]interface{}{},
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
