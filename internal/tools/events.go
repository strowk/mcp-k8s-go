package tools

import (
	"context"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListEventsTool(pool k8s.ClientPool) fxctx.Tool {
	schema := toolinput.NewToolInputSchema(
		toolinput.WithRequiredString("context", "Name of the Kubernetes context to use"),
		toolinput.WithRequiredString("namespace", "Name of the namespace to list events from"),
		toolinput.WithNumber("limit", "Maximum number of events to list"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-events",
			Description: utils.Ptr("List Kubernetes events using specific context in a specified namespace"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(ctx context.Context, args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}

			k8sCtx, err := input.String("context")
			if err != nil {
				return errResponse(err)
			}

			k8sNamespace, err := input.String("namespace")
			if err != nil {
				return errResponse(err)
			}

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return errResponse(err)
			}

			options := metav1.ListOptions{}
			if limit, err := input.Number("limit"); err == nil {
				options.Limit = int64(limit)
			}

			events, err := clientset.
				CoreV1().
				Events(k8sNamespace).
				List(ctx, options)
			if err != nil {
				return errResponse(err)
			}

			var contents = make([]interface{}, len(events.Items))
			for i, event := range events.Items {
				eventInList := EventInList{
					Action:  event.Action,
					Message: event.Message,
					Type:    event.Type,
					Reason:  event.Reason,
					InvolvedObject: InvolvedObject{
						Kind: event.InvolvedObject.Kind,
						Name: event.InvolvedObject.Name,
					},
				}
				content, err := NewJsonContent(eventInList)
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

type InvolvedObject struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type EventInList struct {
	Action         string         `json:"action"`
	Message        string         `json:"message"`
	Type           string         `json:"type"`
	Reason         string         `json:"reason"`
	InvolvedObject InvolvedObject `json:"involvedObject"`
}
