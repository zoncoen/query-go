package yaml

import (
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/zoncoen/query-go"
)

var (
	mapSliceType = reflect.TypeOf(yaml.MapSlice{})
)

// MapSliceExtractFunc is a function for query.CustomExtractFunc option.
func MapSliceExtractFunc(caseInsensitive bool) func(query.ExtractFunc) query.ExtractFunc {
	return func(f query.ExtractFunc) query.ExtractFunc {
		return func(v reflect.Value) (reflect.Value, bool) {
			switch v.Kind() {
			case reflect.Slice:
				if v.Type() == mapSliceType {
					if v.CanInterface() {
						s, ok := v.Interface().(yaml.MapSlice)
						if ok {
							return f(reflect.ValueOf(&keyExtractor{
								v:               s,
								caseInsensitive: caseInsensitive,
							}))
						}
					}
				}
			}
			return f(v)
		}
	}
}

type keyExtractor struct {
	v               yaml.MapSlice
	caseInsensitive bool
}

// ExtractByKey implements the query.KeyExtractor interface.
func (e *keyExtractor) ExtractByKey(key string) (interface{}, bool) {
	if e.caseInsensitive {
		key = strings.ToLower(key)
	}
	for _, i := range e.v {
		k, ok := i.Key.(string)
		if ok {
			if e.caseInsensitive {
				k = strings.ToLower(k)
			}
			if key == k {
				return i.Value, true
			}
		}
	}
	return nil, false
}
