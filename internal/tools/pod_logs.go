package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPodLogsTool(pool k8s.ClientPool) fxctx.Tool {
	schema := toolinput.NewToolInputSchema(
		toolinput.WithRequiredString("context", "Name of the Kubernetes context to use"),
		toolinput.WithRequiredString("namespace", "Name of the namespace where the pod is located"),
		toolinput.WithRequiredString("pod", "Name of the pod to get logs from"),
		toolinput.WithString("sinceDuration", "Only return logs newer than a relative duration like 5s, 2m, or 3h. Only one of sinceTime or sinceDuration may be set."),
		toolinput.WithString("sinceTime", "Only return logs after a specific date (RFC3339). Only one of sinceTime or sinceDuration may be set."),
		toolinput.WithBoolean("previousContainer", "Return previous terminated container logs, defaults to false."),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "get-k8s-pod-logs",
			Description: utils.Ptr("Get logs for a Kubernetes pod using specific context in a specified namespace"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(fmt.Errorf("invalid input: %w", err))
			}

			k8sCtx, err := input.String("context")
			if err != nil {
				return errResponse(fmt.Errorf("invalid input: %w", err))
			}

			k8sNamespace, err := input.String("namespace")
			if err != nil {
				return errResponse(fmt.Errorf("invalid input: %w", err))
			}

			k8sPod, err := input.String("pod")
			if err != nil {
				return errResponse(fmt.Errorf("invalid input: %w", err))
			}

			sinceDurationStr := input.StringOr("sinceDuration", "")

			sinceTimeStr := ""

			sinceTimeStr = input.StringOr("sinceTime", "")

			if sinceDurationStr != "" && sinceTimeStr != "" {
				return errResponse(fmt.Errorf("only one of sinceDuration or sinceTime may be set"))
			}

			previousContainer := input.BooleanOr("previousContainer", false)

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
				Do(ctx)

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

			return &mcp.CallToolResult{
				Content: []interface{}{content},
				IsError: utils.Ptr(false),
			}
		},
	)
}
