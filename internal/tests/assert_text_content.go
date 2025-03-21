package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strowk/foxy-contexts/pkg/mcp"
)

// AssertTextContentContainsInFirstString asserts that the first element of the content is a TextContent and that it contains the expected string
func AssertTextContentContainsInFirstString(t *testing.T, expected string, content []any) {
	if assert.Len(t, content, 1) && assert.IsType(t, mcp.TextContent{}, content[0]) {
		assert.Contains(t, content[0].(mcp.TextContent).Text, "fake logs")
	}
}
