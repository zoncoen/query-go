# query-go

This is a Go package to extract value by a query string like `key[0].key["key"]`.
See usage and example in [GoDoc](https://godoc.org/github.com/zoncoen/query-go).

## Basic Usage

`ParseString` parses a query string and returns the query which extracts the value.

```go
q, err := query.ParseString(`key[0].key["key"]`)
v, err := q.Extract(target)
```

## Query Syntax

The query syntax understood by this package when parsing is as follows.

```txt
.key        extracts by a key of map or field name of struct ("." can be omitted if the head of query)
[0]         extracts by a index of array or slice
["key"]     same as the ".key"
```

