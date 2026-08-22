/*
Package query provides to extract the element from a Go value.

ParseString parses a query string and returns the query which extracts the value.

	q, err := query.ParseString(`$.key[0].key['key']`)
	v, err := q.Extract(ctx, target)

# Query Syntax

The query syntax understood by this package when parsing is as follows.

	$           the root element
	.key        extracts by a key of map or field name of struct ("." can be omitted if the head of query)
	['key']     same as the ".key" (if the key contains "\" or "'", these characters must be escaped like "\\", "\'")
	[0]         extracts by an index of array or slice (a negative index counts from the end: [-1] is the last element), or by an integer key of map
*/
package query
