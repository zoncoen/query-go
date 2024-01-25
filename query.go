package query

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Query represents a query to extract the element from a value.
type Query struct {
	extractors                  []Extractor
	caseInsensitive             bool
	structTags                  []string
	customExtractFuncs          []func(ExtractFunc) ExtractFunc
	customStructFieldNameGetter func(f reflect.StructField) string
	customIsInlineFuncs         []func(reflect.StructField) bool
	hasExplicitRoot             bool
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
func (q Query) Append(es ...Extractor) *Query {
	extractors := make([]Extractor, 0, len(q.extractors)+len(es))
	extractors = append(extractors, q.extractors...)
	extractors = append(extractors, es...)
	q.extractors = extractors
	return &q
}

// Root marks that q has an explicit root operator $.
func (q Query) Root() *Query {
	q.hasExplicitRoot = true
	return &q
}

// Key is shorthand method to create Key and appends it.
func (q Query) Key(k string) *Query {
	return q.Append(&Key{
		key:                k,
		caseInsensitive:    q.caseInsensitive,
		structTags:         q.structTags,
		customExtractFuncs: q.customExtractFuncs,
		fieldNameGetter:    q.customStructFieldNameGetter,
		isInlineFuncs:      q.customIsInlineFuncs,
	})
}

// Index is shorthand method to create Index and appends it.
func (q Query) Index(i int) *Query {
	return q.Append(&Index{index: i})
}

// Extract extracts the value by q from target.
func (q *Query) Extract(target interface{}) (interface{}, error) {
	if q == nil || len(q.extractors) == 0 {
		return target, nil
	}
	v := reflect.ValueOf(target)
	for _, e := range q.extractors {
		f := e.Extract
		for i := len(q.customExtractFuncs) - 1; i >= 0; i-- {
			f = q.customExtractFuncs[i](f)
		}
		var ok bool
		v, ok = f(v)
		if !ok {
			return nil, errors.Errorf(`"%s" not found`, q.String())
		}
		if v.IsValid() && !v.CanInterface() {
			return nil, errors.Errorf("%s: can not access unexported field or method", q.String())
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
	if q.hasExplicitRoot {
		b.WriteString("$")
	}
	for _, f := range q.extractors {
		b.WriteString(f.String())
	}
	return b.String()
}

// Extractors returns query extractors of q.
func (q *Query) Extractors() []Extractor {
	return q.extractors
}

// An Extractor interface is used by a query to extract the element from a value.
type Extractor interface {
	Extract(v reflect.Value) (reflect.Value, bool)
	String() string
}

// ExtractFunc represents a function to extracts a value.
type ExtractFunc func(v reflect.Value) (reflect.Value, bool)
