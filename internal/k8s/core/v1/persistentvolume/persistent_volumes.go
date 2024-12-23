package persistentvolume

import (
	"context"
	"fmt"
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

func NewListPersistentVolumesTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-persistent-volumes",
			Description: utils.Ptr("List Kubernetes Persistent Volumes with detailed information"),
			InputSchema: schema.GetMcpToolInputSchema(),
		},
		func(args map[string]interface{}) *mcp.CallToolResult {
			input, err := schema.Validate(args)
			if err != nil {
				return utils.ErrResponse(err)
			}

			k8sCtx := input.StringOr(contextProperty, "")

			clientset, err := pool.GetClientset(k8sCtx)
			if err != nil {
				return utils.ErrResponse(err)
			}

			pvs, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(pvs.Items, func(i, j int) bool {
				return pvs.Items[i].Name < pvs.Items[j].Name
			})

			var contents []interface{} = make([]interface{}, len(pvs.Items))
			for i, pv := range pvs.Items {
				// Calculate age
				age := time.Since(pv.CreationTimestamp.Time)

				// Determine capacity
				capacity := pv.Spec.Capacity[v1.ResourceStorage]

				// Determine status and claim
				status := string(pv.Status.Phase)
				claimRef := "N/A"
				if pv.Spec.ClaimRef != nil {
					claimRef = fmt.Sprintf("%s/%s",
						pv.Spec.ClaimRef.Namespace,
						pv.Spec.ClaimRef.Name,
					)
				}

				content, err := content.NewJsonContent(PersistentVolumeDetails{
					Name:       pv.Name,
					Capacity:   capacity.String(),
					Status:     status,
					ClaimRef:   claimRef,
					AccessMode: getAccessModes(pv.Spec.AccessModes),
					Age:        utils.FormatAge(age),
					CreatedAt:  pv.CreationTimestamp.Format(time.RFC3339),
					Type:       getPVType(&pv),
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

// PersistentVolumeDetails provides a structured representation of PV information
type PersistentVolumeDetails struct {
	Name       string `json:"name"`
	Capacity   string `json:"capacity"`
	Status     string `json:"status"`
	ClaimRef   string `json:"claimRef"`
	AccessMode string `json:"accessMode"`
	Age        string `json:"age"`
	CreatedAt  string `json:"createdAt"`
	Type       string `json:"type"`
}
