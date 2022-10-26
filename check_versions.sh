#!/bin/bash

git fetch --tags
file_ver=$(awk -F'"' '/APP_VERSION/ {print $2}' main.go)
ch_ver=$(awk -F'[][]' '/\[v[0-9.]+\]/ {print $2; exit}' CHANGELOG.md)
git_tags=$(git describe --tags --always $(git rev-list --tags))

echo "Latest git tag:    ${git_tags%%$'\n'*}"
echo "File version:      ${file_ver}"
echo "Changelog version: ${ch_ver}"
echo

if [[ $file_ver != $ch_ver || $git_tags =~ $file_ver ]]; then
  echo "Version issue!"
  exit 1
else
  echo "Versions look good!"  
fi
