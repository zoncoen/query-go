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
	if v.IsValid() {
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
	case reflect.Slice:
		i := e.index
		if v.Len() > i {
			return v.Index(i), true
		}
	case reflect.Array:
		i := e.index
		if v.Len() > i {
			return v.Index(i), true
		}
	}
	return reflect.Value{}, false
}

// String returns e as string.
func (e *Index) String() string {
	return fmt.Sprintf("[%d]", e.index)
}
