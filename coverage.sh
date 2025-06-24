#!/bin/bash
set -eux -o pipefail

# Use a local directory for Go coverage data
COVERDIR=".coverdir"
export GOCOVERDIR="$COVERDIR"

# Cleanup any previous coverage data before starting
rm -rf "$COVERDIR" coverage.out
mkdir -p "$COVERDIR"

echo "Running tests with GOCOVERDIR=$GOCOVERDIR..."
go test -v ./... -count=1

# Merge coverage files using go tool covdata (no gocovmerge)
echo "Merging and generating coverage report..."
go tool covdata textfmt -i="$COVERDIR" -o=coverage.out

echo "Coverage summary:"
go tool cover -func=coverage.out 
