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
- [list_k8s_logs_test.yaml](testdata/with_k3d/list_k8s_logs_test.yaml) - test for listing logs of a pod

These tests are single or multi document YAML files, where each document is a separate test case with name in "case" field, "in" and "out" for input and expected output being jsonrpc2 requests and responses.

In addition to describing test case, you might need to setup some resources in Kubernetes cluster. 
For how that could be done check methods such as `createPod` in [main_test.go](./main_test.go).
You can either use `kubectl` command or `client-go` library to create and wait for initialization of resources before foxytest test runner starts.

