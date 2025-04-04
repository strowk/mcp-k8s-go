package tools

import (
	"context"
	"encoding/json"
	"log"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"

	"k8s.io/client-go/tools/clientcmd/api"
)

func NewListContextsTool() fxctx.Tool {
	return fxctx.NewTool(
		&mcp.Tool{
			Name:        "list-k8s-contexts",
			Description: utils.Ptr("List Kubernetes contexts from configuration files such as kubeconfig"),
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]map[string]interface{}{},
				Required:   []string{},
			},
		},
		func(_ context.Context, args map[string]interface{}) *mcp.CallToolResult {
			ctx := k8s.GetKubeConfig()
			cfg, err := ctx.RawConfig()
			if err != nil {
				log.Printf("failed to get kubeconfig: %v", err)
				return &mcp.CallToolResult{
					IsError: utils.Ptr(true),
					Meta: map[string]interface{}{
						"error": err.Error(),
					},
					Content: []interface{}{},
				}
			}

			return &mcp.CallToolResult{
				Meta:    map[string]interface{}{},
				Content: getListContextsToolContent(cfg, cfg.CurrentContext),
				IsError: utils.Ptr(false),
			}
		},
	)
}

func getListContextsToolContent(cfg api.Config, current string) []interface{} {
	// First count allowed contexts to allocate the right size
	allowedContextsCount := 0
	for name := range cfg.Contexts {
		if k8s.IsContextAllowed(name) {
			allowedContextsCount++
		}
	}

	var contents = make([]interface{}, allowedContextsCount)
	i := 0

	for name, c := range cfg.Contexts {
		if k8s.IsContextAllowed(name) {
			marshalled, err := json.Marshal(ContextJsonEncoded{
				Context: c,
				Name:    c.Cluster,
				Current: name == current,
			})
			if err != nil {
				log.Printf("failed to marshal context: %v", err)
				continue
			}
			contents[i] = mcp.TextContent{
				Type: "text",
				Text: string(marshalled),
			}

			i++
		}
	}
	return contents
}

type ContextJsonEncoded struct {
	Context *api.Context `json:"context"`
	Name    string       `json:"name"`
	Current bool         `json:"current"`
}
