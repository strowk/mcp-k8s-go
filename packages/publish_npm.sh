#!/bin/bash

set -e

# Mac
mkdir -p ./packages/npm-mcp-k8s-darwin-x64/bin
cp dist/mcp-k8s-go_darwin_amd64_v1/mcp-k8s-go ./packages/npm-mcp-k8s-darwin-x64/bin/mcp-k8s-go
chmod +x ./packages/npm-mcp-k8s-darwin-x64/bin/mcp-k8s-go
mkdir -p ./packages/npm-mcp-k8s-darwin-arm64/bin
cp dist/mcp-k8s-go_darwin_arm64_v8.0/mcp-k8s-go ./packages/npm-mcp-k8s-darwin-arm64/bin/mcp-k8s-go
chmod +x ./packages/npm-mcp-k8s-darwin-arm64/bin/mcp-k8s-go

# Linux
mkdir -p ./packages/npm-mcp-k8s-linux-x64/bin
cp dist/mcp-k8s-go_linux_amd64_v1/mcp-k8s-go ./packages/npm-mcp-k8s-linux-x64/bin/mcp-k8s-go
chmod +x ./packages/npm-mcp-k8s-linux-x64/bin/mcp-k8s-go
mkdir -p ./packages/npm-mcp-k8s-linux-arm64/bin
cp dist/mcp-k8s-go_linux_arm64_v8.0/mcp-k8s-go ./packages/npm-mcp-k8s-linux-arm64/bin/mcp-k8s-go
chmod +x ./packages/npm-mcp-k8s-linux-arm64/bin/mcp-k8s-go

# Windows
mkdir -p ./packages/npm-mcp-k8s-win32-x64/bin
cp dist/mcp-k8s-go_windows_amd64_v1/mcp-k8s-go.exe ./packages/npm-mcp-k8s-win32-x64/bin/mcp-k8s-go.exe
mkdir -p ./packages/npm-mcp-k8s-win32-arm64/bin
cp dist/mcp-k8s-go_windows_arm64_v8.0/mcp-k8s-go.exe ./packages/npm-mcp-k8s-win32-arm64/bin/mcp-k8s-go.exe

cd packages/npm-mcp-k8s-darwin-x64
npm publish --access public

cd ../npm-mcp-k8s-darwin-arm64
npm publish --access public

cd ../npm-mcp-k8s-linux-x64
npm publish --access public

cd ../npm-mcp-k8s-linux-arm64
npm publish --access public

cd ../npm-mcp-k8s-win32-x64
npm publish --access public

cd ../npm-mcp-k8s-win32-arm64
npm publish --access public

cd ../npm-mcp-k8s
npm publish --access public

cd -