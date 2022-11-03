/*
Package query provides to extract the element from a Go value.

ParseString parses a query string and returns the query which extracts the value.

	q, err := query.ParseString(`$.key[0].key["key"]`)
	v, err := q.Extract(target)

# Query Syntax

The query syntax understood by this package when parsing is as follows.

	$           the root element
	.key        extracts by a key of map or field name of struct ("." can be omitted if the head of query)
	[0]         extracts by a index of array or slice
	["key"]     same as the ".key"
*/
package query
