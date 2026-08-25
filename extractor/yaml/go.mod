module github.com/zoncoen/query-go/extractor/yaml

go 1.21.0

toolchain go1.24.1

require (
	github.com/goccy/go-yaml v1.19.2
	github.com/zoncoen/query-go v1.5.0
)

require github.com/pkg/errors v0.9.1 // indirect

// Requires query-go/v2; use github.com/zoncoen/query-go/extractor/yaml/v2 instead.
// These versions adopted the query-go/v2 types while keeping the v1 module
// path, so any build that resolves them for a query-go v1 consumer fails to
// compile. The range also covers the pseudo-versions of the same commits, and
// v0.3.1 itself, which exists only to carry this retraction.
retract [v0.2.4-0, v0.3.1]
