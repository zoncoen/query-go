package query

import (
	"context"
	"fmt"
	"reflect"
)

// IndexExtractor is the interface that wraps the ExtractByIndex method.
//
// ExtractByIndex extracts the value by index.
// It reports whether the index is found and returns the found value.
type IndexExtractor interface {
	ExtractByIndex(index int) (interface{}, bool)
}

// IndexExtractorContext is the interface that wraps the ExtractByIndex method.
//
// ExtractByIndex extracts the value by index.
// It reports whether the index is found and returns the found value.
type IndexExtractorContext interface {
	ExtractByIndex(ctx context.Context, index int) (any, bool)
}

// Index represents an extractor to access the value by index.
//
// A negative index accesses the sequence from the end: -1 is the last
// element, -2 is the second to last, and so on, following the convention
// of RFC 9535 (JSONPath). An index that remains out of range after this
// normalization is reported as not found.
//
// Note that a value implementing IndexExtractor or IndexExtractorContext
// receives the index as given (possibly negative); handling negative
// indices is up to the implementation.
type Index struct {
	index int
}

// Extract extracts the value from v by index.
// It reports whether the index is found and returns the found value.
//
// If v implements the IndexExtractor interface, this method extracts by calling v.ExtractByIndex.
func (e *Index) Extract(v reflect.Value) (reflect.Value, bool) {
	return e.ExtractContext(context.Background(), v)
}

// ExtractContext extracts the value from v by index, passing ctx to a
// context-aware extractor if v implements one.
//
// If v implements the IndexExtractorContext interface, this method extracts by
// calling v.ExtractByIndex with ctx; otherwise it falls back to IndexExtractor.
func (e *Index) ExtractContext(ctx context.Context, v reflect.Value) (reflect.Value, bool) {
	// CanInterface is required: values obtained from unexported fields are
	// read-only and Interface would panic on them.
	if v.IsValid() && v.CanInterface() {
		if i, ok := v.Interface().(IndexExtractorContext); ok {
			x, ok := i.ExtractByIndex(ctx, e.index)
			return reflect.ValueOf(x), ok
		}
		if i, ok := v.Interface().(IndexExtractor); ok {
			x, ok := i.ExtractByIndex(e.index)
			return reflect.ValueOf(x), ok
		}
	}
	return e.extract(v)
}

func (e *Index) extract(v reflect.Value) (reflect.Value, bool) {
	v = elem(v)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		i := e.index
		if i < 0 {
			// A negative index counts backwards from the end of the sequence,
			// following the convention of RFC 9535 (JSONPath).
			i += v.Len()
		}
		if 0 <= i && i < v.Len() {
			return v.Index(i), true
		}
	}
	return reflect.Value{}, false
}

// String returns e as string.
func (e *Index) String() string {
	return fmt.Sprintf("[%d]", e.index)
}
