#!/bin/bash

set -e

# replace previous version with new version in all .json files in ./packages folder 
find ./packages -type f -name '*.json' -exec  sed -i '' -e 's/0.0.9/0.0.10/g' {} \;

# find ./packages -type f -exec sed -i '' -e 's/0.0.2/0.0.3/g' {} \;

cp dist/mcp-k8s-go_darwin_amd64_v1/mcp-k8s-go ./packages/npm-mcp-k8s-darwin-x64/bin/mcp-k8s-go
cp dist/mcp-k8s-go_darwin_arm64_v8.0/mcp-k8s-go ./packages/npm-mcp-k8s-darwin-arm64/bin/mcp-k8s-go
cp dist/mcp-k8s-go_linux_amd64_v1/mcp-k8s-go ./packages/npm-mcp-k8s-linux-x64/bin/mcp-k8s-go
cp dist/mcp-k8s-go_linux_arm64_v8.0/mcp-k8s-go ./packages/npm-mcp-k8s-linux-arm64/bin/mcp-k8s-go
cp dist/mcp-k8s-go_windows_amd64_v1/mcp-k8s-go.exe ./packages/npm-mcp-k8s-win-x64/bin/mcp-k8s-go.exe
cp dist/mcp-k8s-go_windows_arm64_v8.0/mcp-k8s-go.exe ./packages/npm-mcp-k8s-win-arm64/bin/mcp-k8s-go.exe

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