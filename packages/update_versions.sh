#!/bin/bash

# replace previous version with new version in all .json files in ./packages folder 
find ./packages -type f -name '*.json' -exec  sed -i '' -e 's/0.0.9/0.0.10/g' {} \;

