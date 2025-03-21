package pod

import (
	"context"
	"sort"

	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListPods(clientset kubernetes.Interface, k8sNamespace string) *mcp.CallToolResult {
	pods, err := clientset.
		CoreV1().
		Pods(k8sNamespace).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return utils.ErrResponse(err)
	}

	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})

	var contents []interface{} = make([]interface{}, len(pods.Items))
	for i, pod := range pods.Items {
		content, err := content.NewJsonContent(PodInList{
			Name:      pod.Name,
			Namespace: pod.Namespace,
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
}

type PodInList struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
