# query-go

[![Go Reference](https://pkg.go.dev/badge/github.com/zoncoen/query-go/v2.svg)](https://pkg.go.dev/github.com/zoncoen/query-go/v2)
![coverage](docs/coverage.svg) ![ratio](docs/ratio.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoncoen/query-go)](https://goreportcard.com/report/github.com/zoncoen/query-go)
![LICENSE](https://img.shields.io/github/license/zoncoen/query-go.svg)

This is a Go package to extract element from a Go value by a query string like `$.key[0].key['key']`.
See usage and example in the [API reference](https://pkg.go.dev/github.com/zoncoen/query-go/v2).

## Basic Usage

`ParseString` parses a query string and returns the query which extracts the value.

```go
import query "github.com/zoncoen/query-go/v2"

q, err := query.ParseString(`$.key[0].key['key']`)
v, err := q.Extract(ctx, target)
```

When the queried element is absent, the returned error matches
`query.ErrNotFound` via `errors.Is` (and `errors.As` yields a
`*query.NotFoundError` carrying the failed position); any other error is an
extraction failure reported by an extractor, such as a context cancellation
that interrupted a blocking extractor.

## Migrating from v1

- The module path is `github.com/zoncoen/query-go/v2`.
- `Query.Extract` takes a `context.Context`; `ExtractContext` is gone.
- The extractor interfaces are consolidated: `KeyExtractor` and
  `IndexExtractor` now take a context and return `(any, error)` — return
  `query.ErrNotFound` for an absent element instead of `false`. The
  `...Context` interface variants are gone.
- `ExtractFunc` is `func(ctx context.Context, v reflect.Value) (reflect.Value, error)`.
- Extractors can now report failures distinct from absence: any non-ErrNotFound
  error aborts the extraction and is returned to the caller.
- A type that is not migrated silently stops satisfying the interfaces and
  falls back to reflection-based extraction. Add a compile-time assertion to
  each implementation to catch this:

  ```go
  var _ query.KeyExtractor = (*MyType)(nil)
  ```

## Query Syntax

The query syntax understood by this package when parsing is as follows.

```txt
$           the root element
.key        extracts by a key of map or field name of struct ("." can be omitted if the head of query)
['key']     same as the ".key" (if the key contains "\" or "'", these characters must be escaped like "\\", "\'")
[0]         extracts by an index of array or slice (a negative index counts from the end: [-1] is the last element), or by an integer key of map
```
