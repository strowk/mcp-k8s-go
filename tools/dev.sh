#!/bin/bash

npx @modelcontextprotocol/inspector arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -- mcptee dev.log.yaml go run main.go 