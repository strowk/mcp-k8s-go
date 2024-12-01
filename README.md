# MCP K8S Go

This project is intended as a both MCP server connecting to Kubernetes and a library to build more servers for any custom resources in Kubernetes.

Currently available:
- resource: K8S contexts as read from kubeconfig configurations
- tool: list-k8s-contexts
- tool: list-k8s-pods in a given context and namespace
- tool: list-k8s-events in a given context and namespace
- tool: get-k8s-pod-logs in a given context and namespace

## Example usage with Inspector

To use this MCP server with Inspector you can run it from root folder of this project:

```bash
tools/inspector/run.sh
```

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

Now you should be able to run Claude Desktop and:
- see K8S contexts available to attach to conversation as a resource
- ask Claude to list contexts
- ask Claude to list pods in a given context and namespace
- ask Claude to list events in a given context and namespace
- ask Claude to read logs of a given pod in a given context and namespace

### Demo usage with Claude Desktop

Following chat with Claude Desktop demonstrates how it looks when selected particular context as a resource and then asked to check pod logs for errors in kube-system namespace:

![Claude Desktop](docs/images/claude-desktop-logs.png)


