module github.com/zoncoen/query-go/extractor/protobuf

go 1.23

toolchain go1.23.4

require (
	github.com/zoncoen/query-go v1.5.0
	google.golang.org/protobuf v1.36.12
)

require github.com/pkg/errors v0.9.1 // indirect

// Requires query-go/v2; use github.com/zoncoen/query-go/extractor/protobuf/v2 instead.
// These versions adopted the query-go/v2 types while keeping the v1 module
// path, so any build that resolves them for a query-go v1 consumer fails to
// compile. The range also covers the pseudo-versions of the same commits, and
// v0.2.1 itself, which exists only to carry this retraction.
retract [v0.1.6-0, v0.2.1]
