package tools

import (
	"context"

	"github.com/strowk/mcp-k8s-go/internal/k8s"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"

	v1 "k8s.io/api/core/v1"
)

func NewPodLogsTool(pool k8s.ClientPool) fxctx.Tool {
	return fxctx.NewTool(
		"get-k8s-pod-logs",
		"Get logs for a Kubernetes pod using specific context in a specified namespace",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]map[string]interface{}{
				"context": {
					"type": "string",
				},
				"namespace": {
					"type": "string",
				},
				"pod": {
					"type": "string",
				},
			},
			Required: []string{
				"context",
				"namespace",
				"pod",
			},
		},
		func(args map[string]interface{}) fxctx.ToolResponse {
			k8sCtx := args["context"].(string)
			k8sNamespace := args["namespace"].(string)
			k8sPod := args["pod"].(string)

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return errResponse(err)
			}

			podLogs := clientset.
				CoreV1().
				Pods(k8sNamespace).
				GetLogs(k8sPod, &v1.PodLogOptions{}).
				Do(context.Background())

			err = podLogs.Error()

			if err != nil {
				return errResponse(err)
			}

			data, err := podLogs.Raw()
			if err != nil {
				return errResponse(err)
			}

			content := mcp.TextContent{
				Type: "text",
				Text: string(data),
			}

			return fxctx.ToolResponse{
				Content: []interface{}{content},
			}
		},
	)
}
