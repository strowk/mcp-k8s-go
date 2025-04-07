<h4 align="center">Golang-based MCP server connecting to Kubernetes</h4>

<h1 align="center">
   <img src="docs/images/logo.png" width="180"/>
   <br/>
   MCP K8S Go
</h1>

<p align="center">
  <a href="#features">Features</a> ⚙
  <a href="#browse-with-inspector">Browse With Inspector</a> ⚙
  <a href="#use-with-claude">Use With Claude</a> ⚙
  <a href="https://github.com/strowk/mcp-k8s-go/blob/main/CONTRIBUTING.md">Contributing ↗</a>
</p>

<p align="center">
    <a href="https://github.com/strowk/mcp-k8s-go/actions/workflows/dependabot/dependabot-updates"><img src="https://github.com/strowk/mcp-k8s-go/actions/workflows/dependabot/dependabot-updates/badge.svg"></a>
    <a href="https://github.com/strowk/mcp-k8s-go/actions/workflows/test.yaml"><img src="https://github.com/strowk/mcp-k8s-go/actions/workflows/test.yaml/badge.svg"></a>
	  <a href="https://github.com/strowk/mcp-k8s-go/actions/workflows/golangci-lint.yaml"><img src="https://github.com/strowk/mcp-k8s-go/actions/workflows/golangci-lint.yaml/badge.svg"/></a>
    <a href="https://github.com/strowk/mcp-k8s-go/releases/latest"><img src="https://img.shields.io/github/v/release/strowk/mcp-k8s-go?logo=github&color=22ff22" alt="latest release badeg"></a>
    <a href="https://www.npmjs.com/package/@strowk/mcp-k8s"><img src="https://img.shields.io/npm/dw/@strowk/mcp-k8s?label=NPM downloads" alt="npm downloads badge"></a>
    <a href="https://github.com/strowk/mcp-k8s-go/blob/main/LICENSE"><img src="https://img.shields.io/github/license/strowk/mcp-k8s-go" alt="license badge"></a>
</p>

## Features

MCP 💬 prompt 🗂️ resource 🤖 tool 

- 🗂️🤖 List Kubernetes contexts
- 💬🤖 List Kubernetes namespaces
- 🤖 List and get any Kubernetes resources
  - includes custom mappings for resources like pods, services, deployments, but any resource can be listed and retrieved
- 🤖 List Kubernetes nodes
- 💬 List Kubernetes pods
- 🤖 Get Kubernetes events
- 🤖 Get Kubernetes pod logs
- 🤖 Run command in Kubernetes pod

## Browse With Inspector

To use latest published version with Inspector you can run this:

```bash
npx @modelcontextprotocol/inspector npx @strowk/mcp-k8s
```

, or to use version built from sources, then in root folder of this project:

```bash
tools/inspector/run.sh
```

## Use With Claude

<details><summary><b>
Demo Usage
</b></summary>

Following chat with Claude Desktop demonstrates how it looks when selected particular context as a resource and then asked to check pod logs for errors in kube-system namespace:

![Claude Desktop](docs/images/claude-desktop-logs.png)

</details>

To use this MCP server with Claude Desktop (or any other client) you might need to choose which way of installation to use.

You have multiple options:

|              | <a href="#using-smithery">Smithery</a> | <a href="#using-mcp-get">mcp-get</a> | <a href="#prebuilt-from-npm">Pre-built NPM</a> | <a href="#from-github-releases">Pre-built in Github</a> | <a href="#building-from-source">From sources</a> | <a href="#using-docker">Using Docker</a> |
| ------------ | -------------------------------------- | ------------------------------------ | ---------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------ | ---------------------------------------- |
| Claude Setup | Auto                                   | Auto                                 | Manual                                         | Manual                                                  | Manual                                           | Manual                                   |
| Prerequisite | Node.js                                | Node.js                              | Node.js                                        | None                                                    | Golang                                           | Docker                                   |

### Using Smithery

To install MCP K8S Go for Claude Desktop automatically via [Smithery](https://smithery.ai/server/@strowk/mcp-k8s):

```bash
npx -y @smithery/cli install @strowk/mcp-k8s --client claude
```

### Using mcp-get

To install MCP K8S Go for Claude Desktop automatically via [mcp-get](https://mcp-get.com/packages/%40strowk%2Fmcp-k8s):

```bash
npx @michaellatman/mcp-get@latest install @strowk/mcp-k8s
```

### Manually with prebuilt binaries

#### Prebuilt from npm

Use this if you have npm installed and want to use pre-built binaries:

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

, or using `npx` with any client:

```bash
npx @strowk/mcp-k8s
```

For example for Claude:

```json
{
  "mcpServers": {
    "mcp_k8s": {
      "command": "npx",
      "args": [
        "@strowk/mcp-k8s"
      ]
    }
  }
}
```

#### From GitHub releases

Head to [GitHub releases](https://github.com/strowk/mcp-k8s-go/releases) and download the latest release for your platform.

Unpack the archive, which would contain binary named `mcp-k8s-go`, put that binary somewhere in your PATH and then add the following configuration to the `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "mcp_k8s": {
      "command": "mcp-k8s-go",
      "args": []
    }
  }
}
```

### Building from source

You would need Golang installed to build this project:

```bash
go get github.com/strowk/mcp-k8s-go
go install github.com/strowk/mcp-k8s-go
```

, and then add the following configuration to the `claude_desktop_config.json` file:

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

### Using Docker

This server is built and published to Docker Hub since 0.3.1-beta.2 release with multi-arch images available for linux/amd64 and linux/arm64 architectures.

You can use latest tag f.e like this:

```bash
docker run -i -v ~/.kube/config:/root/.kube/config --rm mcpk8s/server:latest
```

Windows users might need to replace `~/.kube/config` with `//c/Users/<username>/.kube/config` at least in Git Bash.

For Claude:

```json
{
  "mcpServers": {
    "mcp_k8s_go": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "-v",
        "~/.kube/config:/root/.kube/config",
        "--rm",
        "mcpk8s/server:latest"
      ]
    }
  }
}
```

### Environment Variables and Command-line Options

The following environment variables are used by the MCP server:

- `KUBECONFIG`: Path to your Kubernetes configuration file (optional, defaults to ~/.kube/config)

The following command-line options are supported:

- `--allowed-contexts=<ctx1,ctx2,...>`: Comma-separated list of allowed Kubernetes contexts that users can access. If not specified, all contexts are allowed.
- `--help`: Display help information
- `--version`: Display version information
