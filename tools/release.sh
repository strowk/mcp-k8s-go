#!/bin/bash

set -e

new_version="${1}"

if [ -z "$new_version" ]; then
  echo "Usage: $0 <new_version>"
  exit 1
fi

# check that new version is X.Y.Z
if [[ ! $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Version should be in format X.Y.Z where X, Y, Z are numbers"
  exit 1
fi

packages/update_versions.sh $new_version
git add ./packages
git commit -m "chore: update npm packages versions to $new_version"

git tag -a "v$new_version" -m "release v$new_version"
goreleaser release --clean

git push origin "v$new_version"


