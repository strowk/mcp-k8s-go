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

func NewListDeploymentsPrompt(pool k8s.ClientPool) fxctx.Prompt {
	return fxctx.NewPrompt(
		mcp.Prompt{
			Name: "list-k8s-deployments",
			Description: utils.Ptr(
				"List Kubernetes Deployments with name and namespace in the current context",
			),
			Arguments: []mcp.PromptArgument{
				{
					Name: "namespace",
					Description: utils.Ptr(
						"Namespace to list Deployments from, defaults to all namespaces",
					),
					Required: utils.Ptr(false),
				},
			},
		},
		func(req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			k8sNamespace := req.Params.Arguments["namespace"]
			if k8sNamespace == "" {
				k8sNamespace = metav1.NamespaceAll
			}

			clientset, err := pool.GetClientset("")
			if err != nil {
				return nil, fmt.Errorf("failed to get k8s client: %w", err)
			}

			deployments, err := clientset.
				AppsV1().
				Deployments(k8sNamespace).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list deployments: %w", err)
			}

			sort.Slice(deployments.Items, func(i, j int) bool {
				return deployments.Items[i].Name < deployments.Items[j].Name
			})

			namespaceInMessage := "all namespaces"
			if k8sNamespace != metav1.NamespaceAll {
				namespaceInMessage = fmt.Sprintf("namespace '%s'", k8sNamespace)
			}

			var messages []mcp.PromptMessage = make(
				[]mcp.PromptMessage,
				len(deployments.Items)+1,
			)
			messages[0] = mcp.PromptMessage{
				Content: mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf(
						"There are %d deployments in %s:",
						len(deployments.Items),
						namespaceInMessage,
					),
				},
				Role: mcp.RoleUser,
			}

			type DeploymentInList struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				Replicas  int32  `json:"replicas"`
			}

			for i, deployment := range deployments.Items {
				content, err := content.NewJsonContent(DeploymentInList{
					Name:      deployment.Name,
					Namespace: deployment.Namespace,
					Replicas:  *deployment.Spec.Replicas,
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
					fmt.Sprintf("Deployments in %s%s", namespaceInMessage, ofContextMsg),
				),
				Messages: messages,
			}, nil
		},
	).WithCompleter(func(arg *mcp.PromptArgument, value string) (*mcp.CompleteResult, error) {
		if arg.Name == "namespace" {
			client, err := pool.GetClientset("")

			if err != nil {
				return nil, fmt.Errorf("failed to get k8s client: %w", err)
			}

			namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
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
