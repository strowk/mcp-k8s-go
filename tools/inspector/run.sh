#!/bin/bash

cd tools/inspector
# npx @modelcontextprotocol/inspector "$@"

npx @modelcontextprotocol/inspector go run ../../main.go "$@"
