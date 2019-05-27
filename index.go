package query

import (
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

// Index represents an extractor to access the value by index.
type Index struct {
	index int
}

// Extract extracts the value from v by index.
// It reports whether the index is found and returns the found value.
//
// If v implements the IndexExtractor interface, this method extracts by calling v.ExtractByIndex.
func (e *Index) Extract(v reflect.Value) (reflect.Value, bool) {
	if v.IsValid() {
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
