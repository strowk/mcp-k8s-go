package prompts

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/content"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewListPodsPrompt(pool k8s.ClientPool) fxctx.Prompt {
	return fxctx.NewPrompt(
		mcp.Prompt{
			Name: "list-k8s-pods",
			Description: utils.Ptr(
				"List Kubernetes Pods with name and namespace in the current context",
			),
			Arguments: []mcp.PromptArgument{
				{
					Name: "namespace",
					Description: utils.Ptr(
						"Namespace to list Pods from, defaults to all namespaces",
					),
					Required: utils.Ptr(false),
				},
			},
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			k8sNamespace := req.Params.Arguments["namespace"]
			if k8sNamespace == "" {
				k8sNamespace = metav1.NamespaceAll
			}

			clientset, err := pool.GetClientset("")
			if err != nil {
				return nil, fmt.Errorf("failed to get k8s client: %w", err)
			}

			pods, err := clientset.
				CoreV1().
				Pods(k8sNamespace).
				List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list pods: %w", err)
			}

			sort.Slice(pods.Items, func(i, j int) bool {
				return pods.Items[i].Name < pods.Items[j].Name
			})

			namespaceInMessage := "all namespaces"
			if k8sNamespace != metav1.NamespaceAll {
				namespaceInMessage = fmt.Sprintf("namespace '%s'", k8sNamespace)
			}

			var messages = make(
				[]mcp.PromptMessage,
				len(pods.Items)+1,
			)
			messages[0] = mcp.PromptMessage{
				Content: mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf(
						"There are %d pods in %s:",
						len(pods.Items),
						namespaceInMessage,
					),
				},
				Role: mcp.RoleUser,
			}

			type PodInList struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			}

			for i, pod := range pods.Items {
				content, err := content.NewJsonContent(PodInList{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create content: %w", err)
				}
				messages[i+1] = mcp.PromptMessage{
					Content: content,
					Role:    mcp.RoleUser,
				}
			}

			ofContextMsg := ""
			currentContext, err := k8s.GetCurrentContext()
			if err == nil && currentContext != "" {
				ofContextMsg = fmt.Sprintf(", context '%s'", currentContext)
			}

			return &mcp.GetPromptResult{
				Description: utils.Ptr(
					fmt.Sprintf("Pods in %s%s", namespaceInMessage, ofContextMsg),
				),
				Messages: messages,
			}, nil
		},
	).WithCompleter(func(ctx context.Context, arg *mcp.PromptArgument, value string) (*mcp.CompleteResult, error) {
		if arg.Name == "namespace" {

			client, err := pool.GetClientset("")

			if err != nil {
				return nil, fmt.Errorf("failed to get k8s client: %w", err)
			}

			namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get namespaces: %w", err)
			}

			var completions []string
			for _, ns := range namespaces.Items {
				if strings.HasPrefix(ns.Name, value) {
					completions = append(completions, ns.Name)
				}
			}

			return &mcp.CompleteResult{
				Completion: mcp.CompleteResultCompletion{
					HasMore: utils.Ptr(false),
					Total:   utils.Ptr(len(completions)),
					Values:  completions,
				},
			}, nil
		}

		return nil, fmt.Errorf("no such argument to complete for prompt: '%s'", arg.Name)
	})
}
