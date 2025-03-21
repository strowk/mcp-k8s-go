package tools

import (
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/mcp-k8s-go/internal/utils"
)

func errResponse(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: utils.Ptr(true),
		Meta:    map[string]interface{}{},
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
