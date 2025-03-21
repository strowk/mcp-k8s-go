package utils

import "github.com/strowk/foxy-contexts/pkg/mcp"

func ErrResponse(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: Ptr(true),
		Meta:    map[string]interface{}{},
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
