package query

import (
	"context"
	"reflect"
	"strings"
)

// KeyExtractor is the interface that wraps the ExtractByKey method.
//
// ExtractByKey extracts the value by key.
// It reports whether the key is found and returns the found value.
type KeyExtractor interface {
	ExtractByKey(key string) (interface{}, bool)
}

// KeyExtractorContext is the interface that wraps the ExtractByKey method.
//
// ExtractByKey extracts the value by key.
// It reports whether the key is found and returns the found value.
type KeyExtractorContext interface {
	ExtractByKey(ctx context.Context, key string) (any, bool)
}

// Key represents an extractor to access the value by key.
type Key struct {
	key                string
	caseInsensitive    bool
	structTags         []string
	customExtractFuncs []func(ExtractFunc) ExtractFunc
	fieldNameGetter    func(f reflect.StructField) string
	isInlineFuncs      []func(reflect.StructField) bool
}

// Extract extracts the value from v by key.
// It reports whether the key is found and returns the found value.
//
// If v implements the KeyExtractor interface, this method extracts by calling v.ExtractByKey.
func (e *Key) Extract(v reflect.Value) (reflect.Value, bool) {
	if v.IsValid() {
		if i, ok := v.Interface().(KeyExtractorContext); ok {
			ctx := withCaseInsensitive(context.Background(), e.caseInsensitive)
			x, ok := i.ExtractByKey(ctx, e.key)
			return reflect.ValueOf(x), ok
		}
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
			if e.caseInsensitive {
				if strings.ToLower(k.String()) == strings.ToLower(e.key) {
					return v.MapIndex(k), true
				}
			} else {
				if k.String() == e.key {
					return v.MapIndex(k), true
				}
			}
		}
	case reflect.Struct:
		inlines := []int{}
		var unexported *reflect.Value
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Type().FieldByIndex([]int{i})
			fieldNames := []string{}
			var inline bool
			for _, t := range e.structTags {
				if s := field.Tag.Get(t); s != "" {
					name, opts, _ := strings.Cut(s, ",")
					if name != "" {
						fieldNames = append(fieldNames, name)
					}
					for _, o := range strings.Split(opts, ",") {
						if o == "inline" {
							inline = true
							break
						}
					}
				}
			}
			fieldNames = append(fieldNames, e.getFieldName(field))
			for _, name := range fieldNames {
				n, k := name, e.key
				if e.caseInsensitive {
					n, k = strings.ToLower(n), strings.ToLower(k)
				}
				if n == k {
					val := v.FieldByIndex([]int{i})
					if isUnexportedField(val) {
						unexported = &val
					} else {
						return val, true
					}
				}
			}
			if field.Anonymous {
				inline = true
			}
			for _, f := range e.isInlineFuncs {
				f := f
				if f(field) {
					inlines = append(inlines, i)
					break
				}
			}
			if inline {
				inlines = append(inlines, i)
			}
		}
		if len(inlines) > 0 {
			f := e.Extract
			for i := len(e.customExtractFuncs) - 1; i >= 0; i-- {
				f = e.customExtractFuncs[i](f)
			}
			for _, i := range inlines {
				val, ok := f(v.FieldByIndex([]int{i}))
				if ok {
					if isUnexportedField(val) {
						unexported = &val
					} else {
						return val, true
					}
				}
			}
		}
		if unexported != nil {
			return *unexported, true
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

func isUnexportedField(v reflect.Value) bool {
	if v.IsValid() && !v.CanInterface() {
		return true
	}
	return false
}

// String returns e as string.
func (e *Key) String() string {
	for _, ch := range e.key {
		switch ch {
		case '[', '.', '\\', '\'':
			return quote(e.key)
		}
	}
	return "." + e.key
}

func quote(s string) string {
	var b strings.Builder
	b.WriteString("['")
	for _, ch := range s {
		switch ch {
		case '\\', '\'':
			b.WriteRune('\\')
			fallthrough
		default:
			b.WriteRune(ch)
		}
	}
	b.WriteString("']")
	return b.String()
}
