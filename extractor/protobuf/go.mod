module github.com/zoncoen/query-go/extractor/protobuf

go 1.21.6

require (
	github.com/zoncoen/query-go v1.3.0
	github.com/zoncoen/query-go/extractor/protobuf/testdata/gen v0.0.0-00010101000000-000000000000
)

replace github.com/zoncoen/query-go/extractor/protobuf/testdata/gen => ./testdata/gen/

require (
	github.com/pkg/errors v0.9.1 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
