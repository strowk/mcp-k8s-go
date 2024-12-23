package deployment

import (
	"context"
	"time"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func NewRestartDeploymentTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"
	nameProperty := "name"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithRequiredString(namespaceProperty, "Namespace to get Deployment from"),
		toolinput.WithRequiredString(nameProperty, "Name of the Deployment to restart"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "restart-k8s-deployment",
			Description: utils.Ptr("Restart Kubernetes Deployment and get updated Deployment"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace, err := input.String(namespaceProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}
			deploymentName, err := input.String(nameProperty)
			if err != nil {
				return utils.ErrResponse(err)
			}

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			_, err = clientset.AppsV1().Deployments(namespace).Patch(context.Background(), deploymentName, types.StrategicMergePatchType, getPatch(), metav1.PatchOptions{})
			if err != nil {
				return utils.ErrResponse(err)
			}
			c, err := content.NewJsonContent(mcp.TextContent{
				Type: "text",
				Text: "Applied patch to restart Deployment",
			})
			if err != nil {
				return utils.ErrResponse(err)
			}
			return &mcp.CallToolResult{
				Meta:    map[string]interface{}{},
				Content: []interface{}{c},
				IsError: utils.Ptr(false),
			}
		},
	)
}

func getPatch() []byte {
	// this imitates the patch that kubectl would send to the server, see here:
	// https://github.com/kubernetes/kubectl/blob/a8a00dbee173287f6b51288d50ddae47e4145f89/pkg/polymorphichelpers/objectrestarter.go#L32
	return []byte(`{"spec":{"template":{"metadata":{"annotations":{"mcp-k8s-go.str4.io/restartedAt":"` + time.Now().Format(time.RFC3339) + `"}}}}}`)
}
