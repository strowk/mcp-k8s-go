#!/bin/bash

cd tools/inspector
npm install
# npx @modelcontextprotocol/inspector "$@"

npx @modelcontextprotocol/inspector go run ../../main.go "$@"
