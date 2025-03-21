package deployment

import (
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeploymentInList provides a structured representation of Deployment information
type DeploymentInList struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Age               string `json:"age"`
	DesiredReplicas   int    `json:"desired_replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	UpdatedReplicas   int    `json:"updated_replicas"`
	AvailableReplicas int    `json:"available_replicas"`
	CreatedAt         string `json:"created_at"`
}

func (d *DeploymentInList) GetName() string {
	return d.Name
}

func (d *DeploymentInList) GetNamespace() string {
	return d.Namespace
}

func NewDeploymentInList(deployment *appsv1.Deployment) *DeploymentInList {
	// Calculate age
	age := time.Since(deployment.CreationTimestamp.Time)

	// Extract deployment status information
	desiredReplicas := int(*(deployment.Spec.Replicas))
	readyReplicas := deployment.Status.ReadyReplicas
	updatedReplicas := deployment.Status.UpdatedReplicas
	availableReplicas := deployment.Status.AvailableReplicas

	return &DeploymentInList{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Age:               utils.FormatAge(age),
		DesiredReplicas:   desiredReplicas,
		ReadyReplicas:     int(readyReplicas),
		UpdatedReplicas:   int(updatedReplicas),
		AvailableReplicas: int(availableReplicas),
		CreatedAt:         deployment.CreationTimestamp.Format(time.RFC3339),
	}
}

func getDeploymentListMapping() list_mapping.ListMapping {
	return func(u runtime.Unstructured) (list_mapping.ListContentItem, error) {
		dep := appsv1.Deployment{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructuredWithValidation(u.UnstructuredContent(), &dep, false)
		if err != nil {
			return nil, err
		}
		return NewDeploymentInList(&dep), nil
	}
}

type listMappingResolver struct{}

func (l *listMappingResolver) GetListMapping(gvk *schema.GroupVersionKind) list_mapping.ListMapping {
	if gvk.Group == "apps" && gvk.Version == "v1" && gvk.Kind == "Deployment" {
		return getDeploymentListMapping()
	}
	return nil
}

func NewListMappingResolver() list_mapping.ListMappingResolver {
	return &listMappingResolver{}
}
