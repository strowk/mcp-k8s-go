package tools

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPodLogsTool(pool k8s.ClientPool) fxctx.Tool {
	return fxctx.NewTool(
		"get-k8s-pod-logs",
		"Get logs for a Kubernetes pod using specific context in a specified namespace",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]map[string]interface{}{
				"context": {
					"type":        "string",
					"description": "Name of the Kubernetes context to use",
				},
				"namespace": {
					"type":        "string",
					"description": "Name of the namespace where the pod is located",
				},
				"pod": {
					"type":        "string",
					"description": "Name of the pod to get logs from",
				},
				"sinceDuration": {
					"type":        "string",
					"description": "Only return logs newer than a relative duration like 5s, 2m, or 3h, only one of sinceTime or sinceDuration may be set.",
				},
				"sinceTime": {
					"type":        "string",
					"description": "Only return logs after a specific date (RFC3339), only one of sinceTime or sinceDuration may be set.",
				},
				"previousContainer": {
					"type":        "boolean",
					"description": "Return previous terminated container logs, defaults to false.",
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

			sinceDurationStr := ""
			sinceDurationArg := args["sinceDuration"]
			if sinceDurationArg != nil {
				sinceDurationStr = sinceDurationArg.(string)
			}

			sinceTimeStr := ""
			sinceTimeArg := args["sinceTime"]
			if sinceTimeArg != nil {
				sinceTimeStr = sinceTimeArg.(string)
			}

			if sinceDurationStr != "" && sinceTimeStr != "" {
				return errResponse(fmt.Errorf("only one of sinceDuration or sinceTime may be set"))
			}

			previousContainer := false
			previousContainerArg := args["previousContainer"]
			if previousContainerArg != nil {
				// TODO: this all has to be moved somewhere separately
				// preferably into some json schema aware validator in foxy-contexts
				previousContainerStr, ok := previousContainerArg.(string)
				if ok {
					if previousContainerStr != "" {
						previousContainerVal, err := strconv.ParseBool(previousContainerStr)
						if err != nil {
							return errResponse(fmt.Errorf("invalid value of previousContainer: %s, expected to be a boolean, got '%w' trying to parse", previousContainerStr, err))
						}
						previousContainer = previousContainerVal
					}
				} else {
					previousContainerVal, ok := previousContainerArg.(bool)
					if !ok {
						return errResponse(fmt.Errorf("invalid type of previousContainer"))
					}
					previousContainer = previousContainerVal
				}
			}

			options := &v1.PodLogOptions{
				Previous: previousContainer,
			}
			if sinceDurationStr != "" {
				sinceDuration, err := time.ParseDuration(sinceDurationStr)
				if err != nil {
					return errResponse(fmt.Errorf("invalid duration: %s, expected to be in Golang duration format as defined in standard time package", sinceDurationStr))
				}

				options.SinceSeconds = utils.Ptr(int64(sinceDuration.Seconds()))
			} else if sinceTimeStr != "" {
				sinceTime, err := time.Parse(time.RFC3339, sinceTimeStr)
				if err != nil {
					return errResponse(fmt.Errorf("invalid time: '%s', expected to be in RFC3339 format, for example 2024-12-01T19:00:08Z", sinceTimeStr))
				}

				options.SinceTime = &metav1.Time{Time: sinceTime}
			}

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return errResponse(err)
			}

			podLogs := clientset.
				CoreV1().
				Pods(k8sNamespace).
				GetLogs(k8sPod, options).
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
