# MCP K8S Go

This project is intended as a both MCP server connecting to Kubernetes and a library to build more servers for any custom resources in Kubernetes.


## Example usage with Claude Desktop

To use this MCP server with Claude Desktop you would firstly need to install it by running:

```bash
go get github.com/strowk/mcp-k8s-go
go install github.com/strowk/mcp-k8s-go
```

, and then add the following configuration to the `claude-desktop.json` file:

```json
{
    "mcpServers": {
        "mcp_k8s_go": {
            "command": "mcp-k8s-go",
            "args": []
        }
    }
}
```

