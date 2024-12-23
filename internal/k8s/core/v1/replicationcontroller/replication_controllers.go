package replicationcontroller

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

func NewListReplicationControllersTool(pool k8s.ClientPool) fxctx.Tool {
	contextProperty := "context"
	namespaceProperty := "namespace"

	schema := toolinput.NewToolInputSchema(
		toolinput.WithString(contextProperty, "Name of the Kubernetes context to use, defaults to current context"),
		toolinput.WithString(namespaceProperty, "Namespace to list Replication Controllers from, defaults to all namespaces"),
	)

	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-replication-controllers",
			Description: utils.Ptr("List Kubernetes Replication Controllers with detailed information"),
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

			var rcs *v1.ReplicationControllerList
			if namespace == "" {
				// List Replication Controllers from all namespaces
				rcs, err = clientset.CoreV1().ReplicationControllers(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			} else {
				// List Replication Controllers from specific namespace
				rcs, err = clientset.CoreV1().ReplicationControllers(namespace).List(context.Background(), metav1.ListOptions{})
			}

			if err != nil {
				return utils.ErrResponse(err)
			}

			sort.Slice(rcs.Items, func(i, j int) bool {
				// Sort by namespace, then by name
				if rcs.Items[i].Namespace == rcs.Items[j].Namespace {
					return rcs.Items[i].Name < rcs.Items[j].Name
				}
				return rcs.Items[i].Namespace < rcs.Items[j].Namespace
			})

			var contents []interface{} = make([]interface{}, len(rcs.Items))
			for i, rc := range rcs.Items {
				// Calculate age
				age := time.Since(rc.CreationTimestamp.Time)

				// Calculate desired and current replicas
				desiredReplicas := int(*(rc.Spec.Replicas))
				currentReplicas := rc.Status.Replicas
				readyReplicas := rc.Status.ReadyReplicas

				content, err := content.NewJsonContent(ReplicationControllerInList{
					Name:            rc.Name,
					Namespace:       rc.Namespace,
					Age:             utils.FormatAge(age),
					DesiredReplicas: desiredReplicas,
					CurrentReplicas: int(currentReplicas),
					ReadyReplicas:   int(readyReplicas),
					CreatedAt:       rc.CreationTimestamp.Format(time.RFC3339),
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

// ReplicationControllerInList provides a structured representation of Replication Controller information
type ReplicationControllerInList struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Age             string `json:"age"`
	DesiredReplicas int    `json:"desired_replicas"`
	CurrentReplicas int    `json:"current_replicas"`
	ReadyReplicas   int    `json:"ready_replicas"`
	CreatedAt       string `json:"createdAt"`
}
