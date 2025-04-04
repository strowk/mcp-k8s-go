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

In addition to describing test case, you might need to setup some resources in Kubernetes cluster. 
For how that could be done check methods such as `createPod` in [main_test.go](./main_test.go).
You can either use `kubectl` command or `client-go` library to create and wait for initialization of resources before foxytest test runner starts.

## Linting

This project uses [golangci-lint](https://golangci-lint.run/) for linting.
Version in use currently is `v2.0.2`.

Once you have installed linter, you can run it with the following command:

```bash
golangci-lint run
```

## Development

You can run the project with automatic reload if you firstly install arelo:

```bash
go install github.com/makiuchi-d/arelo@latest
```

Then simple 

```bash
arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -- go run main.go
```

would start the project's process, you might need to wait a bit before passing any input, as arelo would be building the project.

Now you could start sending requests to the server, for example this would list available tools:

```json
{ "jsonrpc": "2.0", "method": "tools/list", "id": 1, "params": {} }
```

If you want output to be prettified, you can use `jq` and start the server like this:

```bash
arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -- go run main.go | jq
```

Whenever you change any go files, arelo would automatically rebuild the project and restart the server, you might need to send some empty lines though to know whether it is up.
Once empty lines would result in error response, you would know that server is up.

You can also use it with inspector, for example:

```bash
npx @modelcontextprotocol/inspector arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -- go run main.go 
```

, then open `http://localhost:5173/` and Connect to the server, you would have inspector UI to send requests and see responses, while at the same time having automatic reload of the server on any code changes.

Command (example path for windows) in order to use it from any location (useful for providing to any clients, which are part of another program):

```bash
arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -t C:/work/mcp-k8s-go -- go run -C C:/work/mcp-k8s-go main.go
```
<!-- arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -t C:/work/mcp-k8s-go -- mcp-stdio-dev C:/work/mcp-k8s-go/dev.log.yaml go run -C C:/work/mcp-k8s-go main.go -->
<!-- mcp-stdio-dev test.yaml mcp-k8s-go -->

The only issue with this reload is that server is not aware of initialization, however at the moment it does not really preserve any state between restarts that would be depending on initialization, hence noone notices that.

It does have an effect, though, that since clients typically would cache what tools/prompts/resources are available, adding new one would require actually asking for full restart.

## Git Hooks

Setup git hooks used in project by running:

```bash
git config core.hooksPath .githooks
```
