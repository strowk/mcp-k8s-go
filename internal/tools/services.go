package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListServicesTool(pool k8s.ClientPool) fxctx.Tool {
	schema := toolinput.NewToolInputSchema(
		toolinput.WithString("context", "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithRequiredString("namespace", "Name of the namespace to list events from"),
	)
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-services",
			Description: utils.Ptr("List Kubernetes services using specific context in a specified namespace"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return errResponse(err)
			}

			k8sCtx, err := input.String("context")
			if err != nil {
				if errors.Is(err, toolinput.ErrMissingRequestedProperty) {
					k8sCtx = ""
				} else {
					return errResponse(err)
				}
			}

			k8sNamespace, err := input.String("namespace")
			if err != nil {
				return errResponse(err)
			}

			clientset, err := pool.GetClientset(k8sCtx)

			services, err := clientset.
				CoreV1().
				Services(k8sNamespace).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return errResponse(err)
			}

			sort.Slice(services.Items, func(i, j int) bool {
				return services.Items[i].Name < services.Items[j].Name
			})

			var contents []interface{} = make([]interface{}, len(services.Items))
			for i, service := range services.Items {
				serviceContent := ServiceContent{
					Name:        service.Name,
					Namespace:   service.Namespace,
					Type:        string(service.Spec.Type),
					ClusterIP:   service.Spec.ClusterIP,
					ExternalIPs: service.Spec.ExternalIPs,
					Ports:       []string{},
				}

				for _, port := range service.Spec.Ports {
					// this is done similarly to kubectl get services
					// see https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#-em-service-em-
					serviceContent.Ports = append(serviceContent.Ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
				}

				content, err := NewJsonContent(serviceContent)

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

type ServiceContent struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
}
