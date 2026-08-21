package query

import (
	"context"
	"fmt"
	"reflect"
)

// IndexExtractor is the interface that wraps the ExtractByIndex method.
//
// ExtractByIndex extracts the value by index. It returns the found value, or
// ErrNotFound (optionally wrapped) when the index is absent; any other error
// is treated as an extraction failure and aborts the whole query.
type IndexExtractor interface {
	ExtractByIndex(ctx context.Context, index int) (any, error)
}

// Index represents an extractor to access the value by index.
//
// For slices and arrays, a negative index accesses the sequence from the
// end: -1 is the last element, -2 is the second to last, and so on,
// following the convention of RFC 9535 (JSONPath). An index that remains
// out of range after this normalization is reported as absent.
//
// For maps with an integer-kinded key type (and interface-keyed maps
// holding int keys), the index is looked up as the literal map key with no
// normalization: maps have no order, so -1 means the key -1. Maps with a
// floating-point key type are deliberately not supported, matching the map
// key policies of protobuf and CEL: key lookup by equality is unreliable
// for floating-point values.
//
// Note that a value implementing IndexExtractor receives the index as given
// (possibly negative); handling negative indices is up to the
// implementation.
type Index struct {
	index int
}

// Extract extracts the value from v by index, passing ctx to
// v.ExtractByIndex if v implements the IndexExtractor interface. It returns
// ErrNotFound (possibly wrapped) when the index is absent.
func (e *Index) Extract(ctx context.Context, v reflect.Value) (reflect.Value, error) {
	// CanInterface is required: values obtained from unexported fields are
	// read-only and Interface would panic on them.
	if v.IsValid() && v.CanInterface() {
		if i, ok := v.Interface().(IndexExtractor); ok {
			x, err := i.ExtractByIndex(ctx, e.index)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(x), nil
		}
	}
	return e.extract(v)
}

func (e *Index) extract(v reflect.Value) (reflect.Value, error) {
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
			return v.Index(i), nil
		}
	case reflect.Map:
		// An index into a map is the literal map key: maps have no order,
		// so negative indices are not normalized.
		key, ok := e.mapKey(v.Type().Key())
		if !ok {
			break
		}
		if x := v.MapIndex(key); x.IsValid() {
			return x, nil
		}
	}
	return reflect.Value{}, ErrNotFound
}

// mapKey converts the index to a map key of type kt. It reports false when
// kt cannot hold the index: a non-integer key type, or an integer type whose
// range the index does not fit in (a silent Convert would truncate and match
// the wrong key).
func (e *Index) mapKey(kt reflect.Type) (reflect.Value, bool) {
	switch kt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if reflect.New(kt).Elem().OverflowInt(int64(e.index)) {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(e.index).Convert(kt), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if e.index < 0 || reflect.New(kt).Elem().OverflowUint(uint64(e.index)) {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(e.index).Convert(kt), true
	case reflect.Interface:
		if kt.NumMethod() == 0 {
			// Interface-keyed maps are looked up with an int key; keys
			// stored as other integer types are not matched, since
			// interface equality includes the dynamic type.
			return reflect.ValueOf(e.index), true
		}
	}
	return reflect.Value{}, false
}

// String returns e as string.
func (e *Index) String() string {
	return fmt.Sprintf("[%d]", e.index)
}
