package tools

import (
	"context"
	"mcp-k8s-go/internal/k8s"
	"mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListPodsTool(pool k8s.ClientPool) fxctx.Tool {
	return fxctx.NewTool(
		"list-k8s-pods",
		"List Kubernetes pods using specific context in a specified namespace",
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
			// TODO: figure out how to bind args reflectively
			k8sCtx := args["context"].(string)
			k8sNamespace := args["namespace"].(string)

			// TODO: figure out how to use current context, i.e need to make k8sCtx optional

			clientset, err := pool.GetClientset(k8sCtx)

			pods, err := clientset.
				CoreV1().
				Pods(k8sNamespace).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

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

			return fxctx.ToolResponse{
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

func errResponse(err error) fxctx.ToolResponse {
	return fxctx.ToolResponse{
		IsError: utils.Ptr(true),
		Meta: map[string]interface{}{
			"error": err.Error(),
		},
		Content: []interface{}{},
	}
}
