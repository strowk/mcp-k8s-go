#!/bin/bash

set -e

new_version="${1}"

if [ -z "$new_version" ]; then
  echo "Usage: $0 <new_version>"
  echo "hint: last tag is $(git describe --tags --abbrev=0)"
  exit 1
fi

# drop the v prefix
new_version="${new_version#v}"

# check that new version is X.Y.Z or X.Y.Z-beta.N
if ! echo "$new_version" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+(-beta\.[0-9]+)?$' > /dev/null; then
  echo "Version should be in format X.Y.Z where X, Y, Z are numbers, or X.Y.Z-beta.N"
  exit 1
fi

packages/update_versions.sh $new_version
git add ./packages
git commit -m "chore: update npm packages versions to $new_version" --no-verify && git push --no-verify || true

git tag -a "v$new_version" -m "release v$new_version"
git push --no-verify origin "v$new_version"
