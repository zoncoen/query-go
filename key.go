package query

import (
	"context"
	"errors"
	"reflect"
	"strings"
)

// KeyExtractor is the interface that wraps the ExtractByKey method.
//
// ExtractByKey extracts the value by key. It returns the found value, or
// ErrNotFound (optionally wrapped) when the key is absent; any other error
// is treated as an extraction failure and aborts the whole query.
type KeyExtractor interface {
	ExtractByKey(ctx context.Context, key string) (any, error)
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

// Extract extracts the value from v by key, passing ctx (extended with the
// query options) to v.ExtractByKey if v implements the KeyExtractor
// interface. It returns ErrNotFound (possibly wrapped) when the key is
// absent.
func (e *Key) Extract(ctx context.Context, v reflect.Value) (reflect.Value, error) {
	// CanInterface is required: values obtained from unexported fields are
	// read-only and Interface would panic on them.
	if v.IsValid() && v.CanInterface() {
		if i, ok := v.Interface().(KeyExtractor); ok {
			x, err := i.ExtractByKey(withCaseInsensitive(ctx, e.caseInsensitive), e.key)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(x), nil
		}
	}
	return e.extract(ctx, v)
}

func (e *Key) extract(ctx context.Context, v reflect.Value) (reflect.Value, error) {
	v = elem(v)
	switch v.Kind() {
	case reflect.Map:
		if kt := v.Type().Key(); kt.Kind() == reflect.String {
			// Fast path: an exact match is a map lookup, not a linear scan.
			// It also takes precedence over case-insensitive matches.
			if x := v.MapIndex(reflect.ValueOf(e.key).Convert(kt)); x.IsValid() {
				return x, nil
			}
			if !e.caseInsensitive {
				return reflect.Value{}, ErrNotFound
			}
		}
		// Track the smallest matching key so that a case-insensitive lookup
		// is deterministic: MapKeys returns keys in a random order.
		var found reflect.Value
		var foundKey string
		lowerKey := strings.ToLower(e.key)
		for _, k := range v.MapKeys() {
			k := elem(k)
			if k.Kind() != reflect.String {
				// A non-string key can never match: String would return a
				// "<T Value>" placeholder instead of the key itself.
				continue
			}
			ks := k.String()
			if ks == e.key {
				return v.MapIndex(k), nil
			}
			if !e.caseInsensitive {
				continue
			}
			if strings.ToLower(ks) == lowerKey {
				if !found.IsValid() || ks < foundKey {
					found = v.MapIndex(k)
					foundKey = ks
				}
			}
		}
		if found.IsValid() {
			return found, nil
		}
	case reflect.Struct:
		inlines := []int{}
		var unexported *reflect.Value
		for i := range v.Type().NumField() {
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
						return val, nil
					}
				}
			}
			if field.Anonymous {
				inline = true
			}
			for _, f := range e.isInlineFuncs {
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
				val, err := f(ctx, v.FieldByIndex([]int{i}))
				if err == nil {
					if isUnexportedField(val) {
						unexported = &val
					} else {
						return val, nil
					}
				} else if !errors.Is(err, ErrNotFound) {
					// A failure inside an inlined field is a failure of the
					// whole lookup, not an absence.
					return reflect.Value{}, err
				}
			}
		}
		if unexported != nil {
			return *unexported, nil
		}
	}
	return reflect.Value{}, ErrNotFound
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
// The result is parseable: keys that would be tokenized differently in the
// selector notation (e.g. an empty key, or a key containing "$" or "]")
// are rendered in the quoted form.
func (e *Key) String() string {
	if e.key == "" {
		return quote(e.key)
	}
	for _, ch := range e.key {
		switch ch {
		case '[', ']', '.', '\\', '\'', '$':
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
