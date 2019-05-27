package query

import (
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/zoncoen/query-go/ast"
	"github.com/zoncoen/query-go/parser"
)

// Parse parses a query string via r and returns the corresponding Query.
func Parse(r io.Reader, opts ...Option) (*Query, error) {
	node, err := parser.NewParser(r).Parse()
	if err != nil {
		return nil, err
	}
	return buildQuery(New(opts...), node)
}

// ParseString parses a query string s and returns the corresponding Query.
func ParseString(s string, opts ...Option) (*Query, error) {
	return Parse(strings.NewReader(s), opts...)
}

func buildQuery(q *Query, node ast.Node) (*Query, error) {
	if q == nil || node == nil {
		return q, nil
	}
	var err error
	switch n := node.(type) {
	case *ast.Selector:
		q, err = buildQuery(q, n.X)
		if err == nil {
			q = q.Key(n.Sel)
		}
	case *ast.Index:
		q, err = buildQuery(q, n.X)
		if err == nil {
			q = q.Index(n.Index)
		}
	default:
		return nil, errors.Errorf("unknown node type: %T", node)
	}
	return q, err
}
