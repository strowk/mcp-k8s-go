package tools

import (
	"context"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListEventsTool(pool k8s.ClientPool) fxctx.Tool {
	return fxctx.NewTool(
		"list-k8s-events",
		"List Kubernetes events using specific context in a specified namespace",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]map[string]interface{}{
				"context": {
					"type": "string",
				},
				"namespace": {
					"type": "string",
				},
			},
			Required: []string{
				"context",
				"namespace",
			},
		},
		func(args map[string]interface{}) fxctx.ToolResponse {
			k8sCtx := args["context"].(string)
			k8sNamespace := args["namespace"].(string)

			clientset, err := pool.GetClientset(k8sCtx)

			events, err := clientset.
				CoreV1().
				Events(k8sNamespace).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

			var contents []interface{} = make([]interface{}, len(events.Items))
			for i, event := range events.Items {
				content, err := NewJsonContent(EventInList{
					Action:  event.Action,
					Message: event.Message,
				})
				if err != nil {
					return errResponse(err)
				}
				contents[i] = content
			}

			return fxctx.ToolResponse{
				Meta:    map[string]interface{}{},
				Content: contents,
				IsError: utils.Ptr(false),
			}
		},
	)
}

type EventInList struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}
