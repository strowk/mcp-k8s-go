package persistentvolume

import (
	"context"
	"sort"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/toolinput"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListPersistentVolumeClaimsTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list PVCs from, defaults to all namespaces"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-persistent-volume-claims",
			Description: utils.Ptr("List Kubernetes Persistent Volume Claims with detailed information"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")
			namespace := input.StringOr(namespaceProperty, "")

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			var pvcs *v1.PersistentVolumeClaimList
			if namespace == "" {
				// List PVCs from all namespaces
				pvcs, err = clientset.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			} else {
				// List PVCs from specific namespace
				pvcs, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
			}

			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(pvcs.Items, func(i, j int) bool {
				// Sort by namespace, then by name
				if pvcs.Items[i].Namespace == pvcs.Items[j].Namespace {
					return pvcs.Items[i].Name < pvcs.Items[j].Name
				}
				return pvcs.Items[i].Namespace < pvcs.Items[j].Namespace
			})

			var contents []interface{} = make([]interface{}, len(pvcs.Items))
			for i, pvc := range pvcs.Items {
				// Calculate age
				age := time.Since(pvc.CreationTimestamp.Time)

				// Determine capacity and status
				capacity := pvc.Spec.Resources.Requests[v1.ResourceStorage]
				status := string(pvc.Status.Phase)

				// Get volume name
				volumeName := pvc.Spec.VolumeName
				if volumeName == "" {
					volumeName = "N/A"
				}

				// Get access modes
				accessModes := getAccessModes(pvc.Spec.AccessModes)

				content, err := content.NewJsonContent(PersistentVolumeClaimDetails{
					Name:         pvc.Name,
					Namespace:    pvc.Namespace,
					Status:       status,
					Volume:       volumeName,
					Capacity:     capacity.String(),
					AccessMode:   accessModes,
					StorageClass: getStorageClassName(&pvc),
					Age:          utils.FormatAge(age),
					CreatedAt:    pvc.CreationTimestamp.Format(time.RFC3339),
				})
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
		},
	)
}

// PersistentVolumeClaimDetails provides a structured representation of PVC information
type PersistentVolumeClaimDetails struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
	Volume       string `json:"volume"`
	Capacity     string `json:"capacity"`
	AccessMode   string `json:"accessMode"`
	StorageClass string `json:"storageClass"`
	Age          string `json:"age"`
	CreatedAt    string `json:"createdAt"`
}

// getStorageClassName retrieves the storage class name for a PVC
func getStorageClassName(pvc *v1.PersistentVolumeClaim) string {
	if pvc.Spec.StorageClassName != nil {
		return *pvc.Spec.StorageClassName
	}
	return "N/A"
}
