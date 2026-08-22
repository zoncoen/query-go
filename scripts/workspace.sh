#!/usr/bin/env sh
# Set up a Go workspace that resolves github.com/zoncoen/query-go/v2 to this
# checkout, so the extractor modules build against the local root module
# (also required while the version they pin is not tagged yet).
# Used by CI and reproducible locally: run from the repository root.
set -eu
go work init ./extractor/yaml ./extractor/protobuf
go work edit -replace github.com/zoncoen/query-go/v2=.
