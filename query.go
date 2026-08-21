package query

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Query represents a query to extract the element from a value.
type Query struct {
	opts                        []Option
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
	// Keep the original options so that OptionsFromContext reproduces the
	// query's configuration without a hand-maintained reconstruction; clone
	// so that the caller cannot mutate them afterwards.
	q := &Query{opts: slices.Clone(opts)}
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
// For slices and arrays, a negative i accesses the sequence from the end
// (-1 is the last element); for integer-keyed maps, i is the literal map key.
// See Index for the exact semantics.
func (q Query) Index(i int) *Query {
	return q.Append(&Index{index: i})
}

// Extract extracts the value by q from target, passing ctx to each
// extractor (e.g. a value implementing KeyExtractor or IndexExtractor).
//
// When the queried element is absent, the returned error is a *NotFoundError
// matching ErrNotFound via errors.Is. Any other error reported by an
// extractor — e.g. a context cancellation that interrupted a blocking
// extractor — aborts the extraction and is returned wrapped with the
// position of the failing extractor.
func (q *Query) Extract(ctx context.Context, target any) (any, error) {
	if q == nil || len(q.extractors) == 0 {
		return target, nil
	}
	// Expose the query's configuration to extractor implementations; see
	// OptionsFromContext.
	ctx = withOptions(ctx, q.opts)
	v := reflect.ValueOf(target)
	for i, e := range q.extractors {
		f := e.Extract
		for j := len(q.customExtractFuncs) - 1; j >= 0; j-- {
			f = q.customExtractFuncs[j](f)
		}
		var err error
		v, err = f(ctx, v)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, &NotFoundError{Query: q.String(), FailedAt: q.prefixString(i + 1), Err: err}
			}
			return nil, fmt.Errorf("%s: %w", q.prefixString(i+1), err)
		}
		if v.IsValid() && !v.CanInterface() {
			return nil, fmt.Errorf("%s: can not access unexported field or method", q.String())
		}
	}
	if !v.IsValid() {
		return nil, nil
	}
	return v.Interface(), nil
}

// String returns q as string.
func (q *Query) String() string {
	return q.prefixString(len(q.extractors))
}

// prefixString returns the string representation of the first n extractors.
func (q *Query) prefixString(n int) string {
	var b strings.Builder
	if q.hasExplicitRoot {
		b.WriteString("$")
	}
	for _, f := range q.extractors[:n] {
		b.WriteString(f.String())
	}
	return b.String()
}

// Extractors returns a copy of the query extractors of q, so that mutating
// the returned slice cannot corrupt q. Note that extractors carry a snapshot
// of the options they were created with, but a query rebuilt from them via
// Append does not: OptionsFromContext inside such a query reflects the new
// query's own options only.
func (q *Query) Extractors() []Extractor {
	return slices.Clone(q.extractors)
}

// An Extractor interface is used by a query to extract the element from a
// value. Extract returns the extracted value, or ErrNotFound (optionally
// wrapped) when the element is absent; any other error is treated as an
// extraction failure and aborts the query.
type Extractor interface {
	Extract(ctx context.Context, v reflect.Value) (reflect.Value, error)
	String() string
}

// ExtractFunc is the type of the extraction function customized by the
// CustomExtractFunc option.
type ExtractFunc func(ctx context.Context, v reflect.Value) (reflect.Value, error)
