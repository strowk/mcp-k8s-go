# MCP K8S

This is a distribution of MCP server connecting to Kubernetes written in Golang and published to npm.

Currently available:

| 💬 prompt | 🗂️ resource | 🤖 tool |

- 🗂️🤖 List Kubernetes contexts
- 💬🤖 List Kubernetes namespaces
- 🤖 List, get, create and modify any Kubernetes resources
  - includes custom mappings for resources like pods, services, deployments
- 🤖 List Kubernetes nodes
- 💬 List Kubernetes pods
- 🤖 Get Kubernetes events
- 🤖 Get Kubernetes pod logs
- 🤖 Run command in Kubernetes pod

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


### Environment Variables and Command-line Options

The following environment variables are used by the MCP server:

- `KUBECONFIG`: Path to your Kubernetes configuration file (optional, defaults to ~/.kube/config)

The following command-line options are supported:

- `--allowed-contexts=<ctx1,ctx2,...>`: Comma-separated list of allowed Kubernetes contexts that users can access. If not specified, all contexts are allowed.
- `--readonly`: Disables any tool which can write changes to the cluster
- `--help`: Display help information
- `--version`: Display version information

For example if you are configuring Claude Desktop, you can add the following configuration to `claude_desktop_config.json` file:

```json
{
    "mcpServers": {
        "mcp_k8s": {
            "command": "mcp-k8s",
            "args": [
                "--allowed-contexts=dev,prod",
                "--readonly"
            ]
        }
    }
}
```

, which would allow only `dev` and `prod` contexts to be used and would disable any tool which can write changes to the cluster.

