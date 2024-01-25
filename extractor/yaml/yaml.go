/*
Package yaml provides a function to extract values from yaml.MapSlice.
*/
package yaml

import (
	"context"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/zoncoen/query-go"
)

var mapSliceType = reflect.TypeOf(yaml.MapSlice{})

// MapSliceExtractFunc is a function for query.CustomExtractFunc option to extract values from yaml.MapSlice.
func MapSliceExtractFunc() func(query.ExtractFunc) query.ExtractFunc {
	return func(f query.ExtractFunc) query.ExtractFunc {
		return func(in reflect.Value) (reflect.Value, bool) {
			v := in
			for {
				if v.IsValid() {
					if k := v.Kind(); k == reflect.Interface || k == reflect.Pointer {
						v = v.Elem()
						continue
					}
				}
				break
			}
			switch v.Kind() {
			case reflect.Slice:
				if v.Type() == mapSliceType {
					if v.CanInterface() {
						s, ok := v.Interface().(yaml.MapSlice)
						if ok {
							return f(reflect.ValueOf(&keyExtractor{
								v:               s,
							}))
						}
					}
				}
			}
			return f(in)
		}
	}
}

type keyExtractor struct {
	v               yaml.MapSlice
}

// ExtractByKey implements the query.KeyExtractorContext interface.
func (e *keyExtractor) ExtractByKey(ctx context.Context, key string) (interface{}, bool) {
	ci := query.IsCaseInsensitive(ctx)
	if ci {
		key = strings.ToLower(key)
	}
	for _, i := range e.v {
		k, ok := i.Key.(string)
		if ok {
			if ci {
				k = strings.ToLower(k)
			}
			if key == k {
				return i.Value, true
			}
		}
	}
	return nil, false
}
