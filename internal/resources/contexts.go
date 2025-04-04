package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"k8s.io/client-go/tools/clientcmd/api"
)

func NewContextsResourceProvider() fxctx.ResourceProvider {
	return fxctx.NewResourceProvider(
		func(_ context.Context) ([]mcp.Resource, error) {
			cfg, err := k8s.GetKubeConfig().RawConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
			}

			resources := []mcp.Resource{}
			for name := range cfg.Contexts {
				if k8s.IsContextAllowed(name) {
					resources = append(resources, toMcpResourcse(name))
				}
			}
			return resources, nil
		},

		func(_ context.Context, uri string) (*mcp.ReadResourceResult, error) {
			cfg, err := k8s.GetKubeConfig().RawConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
			}

			if uri == "contexts" {
				contents := getContextsContent(uri, cfg)
				return &mcp.ReadResourceResult{
					Contents: contents,
				}, nil
			}

			if strings.HasPrefix(uri, "contexts/") {
				name := strings.TrimPrefix(uri, "contexts/")
				c, ok := cfg.Contexts[name]
				if !ok {
					return nil, fmt.Errorf("context not found: %s", name)
				}

				var contents = make([]interface{}, 1)
				contents[0] = &struct {
					Uri     string       `json:"uri"`
					Text    string       `json:"text"`
					Context *api.Context `json:"context"`
					Name    string       `json:"name"`
				}{Context: c, Name: name, Text: name, Uri: uri}

				return &mcp.ReadResourceResult{
					Contents: contents,
				}, nil
			}

			return nil, nil
		})
}

func toMcpResourcse(contextName string) mcp.Resource {
	return mcp.Resource{Annotations: &mcp.ResourceAnnotations{
		Audience: []mcp.Role{mcp.RoleAssistant, mcp.RoleUser},
	},
		Name:        contextName,
		Description: utils.Ptr("Specific k8s context as read from kubeconfig configuration files"),
		Uri:         "contexts/" + contextName,
	}
}

func getContextsContent(uri string, cfg api.Config) []interface{} {
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
			contents[i] = ContextContent{
				Uri:  uri + "/" + name,
				Text: name,

				Context: c,
				Name:    name,
			}
			i++
		}
	}
	return contents
}

type ContextContent struct {
	Uri     string       `json:"uri"`
	Text    string       `json:"text"`
	Context *api.Context `json:"context"`
	Name    string       `json:"name"`
}
