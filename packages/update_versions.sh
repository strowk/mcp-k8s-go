#!/bin/bash

previous_version="$(npm view ./packages/npm-mcp-k8s version)"
new_version="${1}"

if [ -z "$new_version" ]; then
  echo "Usage: $0 <new_version>"
  exit 1
fi

# replace previous version with new version in all .json files in ./packages folder 
find ./packages -type f -name '*.json' -exec  sed -i '' -e "s/${previous_version}/${new_version}/g" {} \;

