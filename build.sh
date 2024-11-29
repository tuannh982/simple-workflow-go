#!/bin/sh

set -eo pipefail

go clean
go clean -testcache
go build ./...
go test ./...
go test -race ./...

if [[ $ENABLE_E2E_TEST ]]; then
  go test -tags e2e -p 1 ./test/e2e/...
fi