package query

import (
	"reflect"
)

// KeyExtractor is the interface that wraps the ExtractByKey method.
//
// ExtractByKey extracts the value by key.
// It reports whether the key is found and returns the found value.
type KeyExtractor interface {
	ExtractByKey(key string) (interface{}, bool)
}

// Key represents an extractor to access the value by key.
type Key struct {
	key             string
	fieldNameGetter func(f reflect.StructField) string
}

// Extract extracts the value from v by key.
// It reports whether the key is found and returns the found value.
//
// If v implements the KeyExtractor interface, this method extracts by calling v.ExtractByKey.
func (e *Key) Extract(v reflect.Value) (reflect.Value, bool) {
	if v.IsValid() {
		if i, ok := v.Interface().(KeyExtractor); ok {
			x, ok := i.ExtractByKey(e.key)
			return reflect.ValueOf(x), ok
		}
	}
	return e.extract(v)
}

func (e *Key) extract(v reflect.Value) (reflect.Value, bool) {
	v = elem(v)
	switch v.Kind() {
	case reflect.Map:
		for _, k := range v.MapKeys() {
			k := elem(k)
			if k.String() == e.key {
				return v.MapIndex(k), true
			}
		}
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Type().FieldByIndex([]int{i})
			if e.getFieldName(field) == e.key {
				return v.FieldByIndex([]int{i}), true
			}
		}
	}
	return reflect.Value{}, false
}

func (e *Key) getFieldName(field reflect.StructField) string {
	if e.fieldNameGetter != nil {
		return e.fieldNameGetter(field)
	}
	return field.Name
}

// String returns e as string.
func (e *Key) String() string {
	return "." + e.key
}
