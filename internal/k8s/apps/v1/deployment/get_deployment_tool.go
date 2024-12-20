package deployment

import (
	"context"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewGetDeploymentTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"
	nameProperty := "name"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithRequiredString(namespaceProperty, "Namespace to get Deployment from"),
		toolinput.WithRequiredString(nameProperty, "Name of the Deployment to get"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "get-k8s-deployment",
			Description: utils.Ptr("Get single Kubernetes Deployment with detailed information"),
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

			var deployment *appsv1.Deployment
			deployment, err = clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
			if err != nil {
				return utils.ErrResponse(err)
			}
			utils.SanitizeObjectMeta(&deployment.ObjectMeta)

			c, err := content.NewJsonContent(deployment)
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
