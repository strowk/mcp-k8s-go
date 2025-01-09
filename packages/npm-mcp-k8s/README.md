# MCP K8S

This is a distribution of MCP server connecting to Kubernetes written in Golang and published to npm.

Currently available:

- 🗂️ resource: K8S contexts as read from kubeconfig configurations
- 🤖 tool: list-k8s-contexts
- 🤖 tool: list-k8s-namespaces in a given context
- 🤖 tool: list-k8s-nodes in a given context
- 🤖 tool: list-k8s-resources in a given context and namespace for specified resource kind
  - includes custom mappings for resources like pods, services, deployments
- 🤖 tool: get-k8s-resource in a given context and namespace for specified name and resource kind
- 🤖 tool: list-k8s-events in a given context and namespace
- 🤖 tool: get-k8s-pod-logs in a given context and namespace for specified pod
- 💬 prompt: list-k8s-namespaces in a given context
- 💬 prompt: list-k8s-pods in current context and with given namespace

## Example usage with Claude Desktop

To use this MCP server with Claude Desktop you would firstly need to install it.

You have two options at the moment - use pre-built binaries or build it from source. Refer to [sources](https://github.com/strowk/mcp-k8s-go/) to know how to build from sources.

### Installing from npm

```bash
npm install -g @strowk/mcp-k8s
```

Then check version by running `mcp-k8s --version` and if this printed installed version, you can proceed to add configuration to `claude_desktop_config.json` file:

```json
{
    "mcpServers": {
        "mcp_k8s": {
            "command": "mcp-k8s",
            "args": []
        }
    }
}
```

### Using from Claude Desktop

Now you should be able to run Claude Desktop and:
- see K8S contexts available to attach to conversation as a resource
- ask Claude to list contexts
- ask Claude to list pods in a given context and namespace
- ask Claude to list events in a given context and namespace
- ask Claude to read logs of a given pod in a given context and namespace

### Demo usage with Claude Desktop

Following chat with Claude Desktop demonstrates how it looks when selected particular context as a resource and then asked to check pod logs for errors in kube-system namespace:

![Claude Desktop](docs/images/claude-desktop-logs.png)


