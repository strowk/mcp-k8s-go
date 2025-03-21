# CLAUDE.md - MCP K8S Go Development Guide

## Build & Test Commands
- Build: `go build`
- Run with hot reload: `arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -- go run main.go`
- Run all tests: `go test`
- Run single test: `go test -run '^TestName$'` (example: `go test -run '^TestListContexts$'`)
- Generate mocks: `tools/generate-mocks.sh`
- Lint: `golangci-lint run`

## Code Style Guidelines
- **Imports**: Standard Go import organization (stdlib, external, internal)
- **Error Handling**: Return errors explicitly; prefer wrapping with context
- **Naming**: Use Go conventions (CamelCase for exported, camelCase for unexported)
- **Testing**: Use YAML test files in testdata directory with foxytest package
- **Types**: Use strong typing; prefer interfaces for dependencies
- **Documentation**: Document all exported functions and types
- **Structure**: Follow k8s-like API structure in internal/k8s package
- **Dependencies**: Use dependency injection with fx framework
- **Context**: Pass kubernetes contexts explicitly as parameters