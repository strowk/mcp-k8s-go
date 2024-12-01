package tools

import (
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
		"list-k8s-contexts",
		"List Kubernetes contexts from configuration files such as kubeconfig",
		mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]map[string]interface{}{},
			Required:   []string{},
		},
		func(args map[string]interface{}) fxctx.ToolResponse {
			ctx := k8s.GetKubeConfig()
			cfg, err := ctx.RawConfig()
			if err != nil {
				log.Printf("failed to get kubeconfig: %v", err)
				return fxctx.ToolResponse{
					IsError: utils.Ptr(true),
					Meta: map[string]interface{}{
						"error": err.Error(),
					},
					Content: []interface{}{},
				}
			}

			return fxctx.ToolResponse{
				Meta:    map[string]interface{}{},
				Content: getListContextsToolContent(cfg),
				IsError: utils.Ptr(false),
			}
		},
	)
}

func getListContextsToolContent(cfg api.Config) []interface{} {
	var contents []interface{} = make([]interface{}, len(cfg.Contexts))
	i := 0

	for _, c := range cfg.Contexts {
		marshalled, err := json.Marshal(ContextJsonEncoded{
			Context: c,
			Name:    c.Cluster,
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
	return contents
}

type ContextJsonEncoded struct {
	Context *api.Context `json:"context"`
	Name    string       `json:"name"`
}
