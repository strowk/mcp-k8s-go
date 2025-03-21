package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceContent struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
}

func ListServices(clientset kubernetes.Interface, k8sNamespace string) *mcp.CallToolResult {
	services, err := clientset.
		CoreV1().
		Services(k8sNamespace).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return utils.ErrResponse(err)
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

		content, err := content.NewJsonContent(serviceContent)

		if err != nil {
			return utils.ErrResponse(err)
		}
		contents[i] = content
	}

	return &mcp.CallToolResult{
		Meta:    map[string]interface{}{},
		Content: contents,
		IsError: utils.Ptr(false),
	}
}
