package service

import (
	"fmt"

	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceContent struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
}

func NewServiceContent(service *corev1.Service) *ServiceContent {
	serviceContent := &ServiceContent{
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

	return serviceContent
}

func (s *ServiceContent) GetName() string {
	return s.Name
}

func (s *ServiceContent) GetNamespace() string {
	return s.Namespace
}

func getServiceListMapping() list_mapping.ListMapping {
	return func(u runtime.Unstructured) (list_mapping.ListContentItem, error) {
		svc := corev1.Service{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructuredWithValidation(u.UnstructuredContent(), &svc, false)
		if err != nil {
			return nil, err
		}
		return NewServiceContent(&svc), nil
	}
}

type listMappingResolver struct {
	list_mapping.ListMappingResolver
}

func (r *listMappingResolver) GetListMapping(gvk *schema.GroupVersionKind) list_mapping.ListMapping {
	if (gvk.Group == "core" || gvk.Group == "") && gvk.Version == "v1" && gvk.Kind == "Service" {
		return getServiceListMapping()
	}
	return nil
}

func NewListMappingResolver() list_mapping.ListMappingResolver {
	return &listMappingResolver{}
}
