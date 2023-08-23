#!/bin/bash

set -e

gofmt -s -d $(find . -type f -name '*.go') |& perl -pe 'END{exit($. > 0 ? 1 : 0)}'

go test -v -race ./...                   # Run all the tests with the race detector enabled

./tests/all.bash

echo "Running extra checks..."
staticcheck ./...

echo "All check passed."