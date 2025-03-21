package deployment

import (
	"context"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDeployment(
	clientset kubernetes.Interface,
	namespace string,
	name string,
	templateStr string,
) *mcp.CallToolResult {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return utils.ErrResponse(err)
	}

	var cnt interface{}
	if templateStr != "" {
		tmpl, err := template.New("template").Parse(templateStr)
		if err != nil {
			return utils.ErrResponse(err)
		}

		// transforming deployment to JSON and back so that it has clear structure
		parsedDeployment, err := json.Marshal(deployment)
		if err != nil {
			return utils.ErrResponse(err)
		}

		var parsedDeploymentMap map[string]interface{}
		err = json.Unmarshal(parsedDeployment, &parsedDeploymentMap)
		if err != nil {
			return utils.ErrResponse(err)
		}

		buf := new(strings.Builder)

		err = tmpl.Execute(buf, parsedDeploymentMap)
		if err != nil {
			return utils.ErrResponse(err)
		}

		cnt = mcp.TextContent{
			Type: "text",
			Text: buf.String(),
		}
	} else {
		utils.SanitizeObjectMeta(&deployment.ObjectMeta)
		c, err := content.NewJsonContent(deployment)
		if err != nil {
			return utils.ErrResponse(err)
		}
		cnt = c
	}

	return &mcp.CallToolResult{
		Meta:    map[string]interface{}{},
		Content: []interface{}{cnt},
		IsError: utils.Ptr(false),
	}
}
