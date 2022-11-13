# query-go

[![GoDoc](https://godoc.org/github.com/zoncoen/query-go?status.svg)](https://godoc.org/github.com/zoncoen/query-go)
[![Build Status](https://travis-ci.org/zoncoen/query-go.svg?branch=main)](https://travis-ci.org/zoncoen/query-go)
[![codecov](https://codecov.io/gh/zoncoen/query-go/branch/main/graph/badge.svg)](https://codecov.io/gh/zoncoen/query-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoncoen/query-go)](https://goreportcard.com/report/github.com/zoncoen/query-go)
![LICENSE](https://img.shields.io/github/license/zoncoen/query-go.svg)

This is a Go package to extract element from a Go value by a query string like `$.key[0].key['key']`.
See usage and example in [GoDoc](https://godoc.org/github.com/zoncoen/query-go).

## Basic Usage

`ParseString` parses a query string and returns the query which extracts the value.

```go
q, err := query.ParseString(`$.key[0].key['key']`)
v, err := q.Extract(target)
```

## Query Syntax

The query syntax understood by this package when parsing is as follows.

```txt
$           the root element
.key        extracts by a key of map or field name of struct ("." can be omitted if the head of query)
['key']     same as the ".key" (if the key contains "\" or "'", these characters must be escaped like "\\", "\'")
[0]         extracts by a index of array or slice
```
