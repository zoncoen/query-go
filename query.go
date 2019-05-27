// Package query provides to extract the element from a value.
package query

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Query represents a query to extract the element from a value.
type Query struct {
	extractors                  []Extractor
	customStructFieldNameGetter func(f reflect.StructField) string
}

// New returns a new query.
func New(opts ...Option) *Query {
	q := &Query{}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// Append appends extractor to q and returns updated q.
func (q Query) Append(extractor Extractor) *Query {
	q.extractors = append(q.extractors, extractor)
	return &q
}

// Key is shorthand method to create Key and appends it.
func (q Query) Key(k string) *Query {
	return q.Append(&Key{
		key:             k,
		fieldNameGetter: q.customStructFieldNameGetter,
	})
}

// Index is shorthand method to create Index and appends it.
func (q Query) Index(i int) *Query {
	return q.Append(&Index{index: i})
}

// Extract extracts the value by q from target.
func (q *Query) Extract(target interface{}) (interface{}, error) {
	if q == nil {
		return target, nil
	}
	v := reflect.ValueOf(target)
	for _, f := range q.extractors {
		var ok bool
		v, ok = f.Extract(v)
		if !ok {
			return nil, errors.Errorf(`"%s" not found`, q.String())
		}
	}
	if !v.IsValid() {
		return nil, nil
	}
	return v.Interface(), nil
}

// String returns q as string.
func (q *Query) String() string {
	var b strings.Builder
	for _, f := range q.extractors {
		b.WriteString(f.String())
	}
	return b.String()
}

// An Extractor interface is used by a query to extract the element from a value.
type Extractor interface {
	Extract(v reflect.Value) (reflect.Value, bool)
	String() string
}
