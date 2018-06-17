#!/bin/bash

set -e

gofmt -s -d $(find . -type f -name '*.go') |& perl -pe 'END{exit($. > 0 ? 1 : 0)}'

go test -v -race ./...                   # Run all the tests with the race detector enabled

./tests/all.bash

echo "Running extra checks..."
#go vet ./... |& grep -v 'method UnmarshalJSON' |& perl -pe 'END{exit($. > 0 ? 1 : 0)}'
megacheck ./... |& grep -v '\(SA9003\|U1000\)' |& perl -pe 'END{exit($. > 0 ? 1 : 0)}'
golint $(go list ./...) |& grep -v '\(never used\|is unused\|comment on exported\|exported .* should have\)' | perl -pe 'END{exit($. > 0 ? 1 : 0)}'

echo "All check passed."