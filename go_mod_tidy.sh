#!/bin/bash

# Loop through all directories in the current folder
for dir in */; do
  # Check if it's a directory and contains a go.mod file
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    cd "$dir" || exit
    go mod tidy
    cd ..
  fi
done
