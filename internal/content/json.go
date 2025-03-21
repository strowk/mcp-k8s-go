package content

import (
	"encoding/json"

	"github.com/strowk/foxy-contexts/pkg/mcp"
)

func NewJsonContent(v any) (mcp.TextContent, error) {
	contents, err := json.Marshal(v)
	if err != nil {
		return mcp.TextContent{}, err
	}
	return mcp.TextContent{
		Type: "text",
		Text: string(contents),
	}, nil
}
