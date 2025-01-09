package deployment

import (
	"context"

	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDeployment(clientset kubernetes.Interface, namespace string, name string) *mcp.CallToolResult {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
}
