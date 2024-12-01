#!/bin/bash

previous_version="0.0.9"
new_version="0.0.10"

# replace previous version with new version in all .json files in ./packages folder 
find ./packages -type f -name '*.json' -exec  sed -i '' -e "s/${previous_version}/${new_version}/g" {} \;

