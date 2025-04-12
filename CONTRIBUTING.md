# Contribution Guidelines

## Contributing

I welcome any contributions to this project. Please follow the guidelines below.

## Issues

If you find a bug or have a feature request, please open an issue in Github. If you are able to provide test to reproduce the issue, that would be very helpful. See further testing guidelines below.

## Coding

This project uses Go modules, so you should be able to clone the repository to any location on your machine.

To build the project, you can run:

```bash
go build
```

## Testing

This project uses [k3d](https://k3d.io/) to run Kubernetes cluster locally for testing.
As code is written in Go, you would also need to have Go installed.

To run tests, you need to have `k3d` and `kubectl` installed and working (for which you also need Docker).

During the test run, a new k3d cluster will be created and then deleted after the tests are finished. In some situations the cluster might not be deleted properly, so you might need to delete it manually by running:

```bash
k3d cluster delete mcp-k8s-integration-test
```

To run tests, execute the following command:

```bash
go test
```

First run might take longer, as some images are being downloaded.
Consequent runs would still not be very fast, as the whole k3d cluster is being created and deleted for each test run.

Some limited amount of tests do not require k3d cluster to be running, for example `TestListContexts` and `TestListTools`.

Here is an example of running only `TestListContexts` test:

```bash
go test -run '^TestListContexts$'
```

### Adding new test

To describe a test case, this project uses foxytest package with every test being a separate YAML document.

Check tests in testdata directory for examples:
- [list_tools_test.yaml](testdata/list_tools_test.yaml) - test for listing tools
- [get_k8s_pod_logs_test.yaml](testdata/with_k3d/get_k8s_pod_logs_test.yaml) - test for listing logs of a pod

These tests are single or multi document YAML files, where each document is a separate test case with name in "case" field, "in" and "out" for input and expected output being jsonrpc2 requests and responses.

For new tests which are related to one particular resource, files should be located under `internal/k8s/<group>/<version>/<resource>` folder, for example `internal/k8s/apps/v1/deployment`. If you create new such folder, then you would need to add it in the list of test suites in `TestInK3dCluster` function in [main_test.go](main_test.go) file.

In addition to describing test case, you might need to setup some resources in Kubernetes cluster.
For this you have to place YAML files describing these resources in test suite subfolder called `test_manifests`. For example when tests within `internal/k8s/apps/v1/deployment` package are run, test manifests should be in `internal/k8s/apps/v1/deployment/test_manifests` folder and would be applied to the cluster before that test suite is run.

## Linting

This project uses [golangci-lint](https://golangci-lint.run/) for linting.
Version in use currently is `v2.0.2`.

Once you have installed linter, you can run it with the following command:

```bash
golangci-lint run
```

## Development

### Hot Reloading Setup With Logging

#### TL;DR

Install [synf](https://github.com/strowk/synf?tab=readme-ov-file#installation) and mcptee:

```bash
go install github.com/strowk/mcptee@latest
```

Then you can use command like this:

```bash
mcptee dev.log.yaml synf dev .
```
, or if you configure this with some client, probably with full path to the project (replace `C:/work/mcp-k8s-go` with path to where project repository is cloned):

```bash
mcptee C:/work/mcp-k8s-go/dev.log.yaml synf dev C:/work/mcp-k8s-go
```

This would be for Claude:

```json
{
    "mcpServers": {
      "mcp_k8s_dev": {
        "command": "mcptee",
        "args": [
          "C:/work/mcp-k8s-go/dev.log.yaml",
          "synf",
          "dev",
          "C:/work/mcp-k8s-go"
        ]
      }
    }
}
```

#### Long Version

You can run the project with automatic reload if you firstly install [synf](https://github.com/strowk/synf?tab=readme-ov-file#installation) tool.

Then simple command

```bash
synf dev .
```

would start the project's process, you might need to wait a bit before passing any input, as synf would be building the project.

Now you could start sending requests to the server, for example this would list available tools:

```json
{ "jsonrpc": "2.0", "method": "tools/list", "id": 1, "params": {} }
```

If you want output to be prettified, you can use `jq` and start the server like this:

```bash
synf dev . | jq
```

Whenever you change any go files, synf would automatically rebuild the project and restart the server, you might need to send some empty lines though to know whether it is up.
Once empty lines would result in error response, you would know that server is up.

You can also use it with inspector, for example:

```bash
npx @modelcontextprotocol/inspector synf dev .
```

, then open url that inspector would print and Connect to the server, you would have inspector UI to send requests and see responses, while at the same time having automatic reload of the server on any code changes.

Command (example path for windows) in order to use it from any location (useful for providing to any clients, which are part of another program):

```bash
mcptee log.yaml synf dev C:/work/mcp-k8s-go
```

Synf would make sure that client receives list_update notification whenever the server is restarted, which should make clients that support this to pick it up automatically.

If you would also like to capture communication between the server and client, you can use `mcptee` tool.
You can install it with `go install github.com/strowk/mcptee@latest` command.

`mcptee` would capture all the communication between the server and client, and write it to a YAML file to use for debugging.

For example: `mcptee log.yaml synf dev .` would start server with hot reloading and logging to `log.yaml` file.

## Git Hooks

Setup git hooks used in project by running:

```bash
git config core.hooksPath .githooks
```
