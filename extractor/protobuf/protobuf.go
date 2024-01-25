package protobuf

import (
	"context"
	"reflect"
	"strings"

	"github.com/zoncoen/query-go"
)

// ExtractFunc is a function for query.CustomExtractFunc option to extract values by protobuf struct tag.
func ExtractFunc() func(query.ExtractFunc) query.ExtractFunc {
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
			case reflect.Struct:
				for i := 0; i < v.Type().NumField(); i++ {
					field := v.Type().FieldByIndex([]int{i})
					if s := field.Tag.Get("protobuf"); s != "" {
						if v, found := f(reflect.ValueOf(&keyExtractor{v})); found {
							return v, true
						}
					}
				}
			}
			return f(in)
		}
	}
}

type keyExtractor struct {
	v reflect.Value
}

// ExtractByKey implements KeyExtractorContext interface.
func (e *keyExtractor) ExtractByKey(ctx context.Context, key string) (any, bool) {
	ci := query.IsCaseInsensitive(ctx)
	if ci {
		key = strings.ToLower(key)
	}
	switch e.v.Kind() {
	case reflect.Struct:
		for i := 0; i < e.v.Type().NumField(); i++ {
			if s := e.v.Type().FieldByIndex([]int{i}).Tag.Get("protobuf"); s != "" {
				for _, opt := range strings.Split(s, ",") {
					kv := strings.Split(opt, "=")
					if len(kv) == 2 {
						k, v := kv[0], kv[1]
						if k == "name" {
							if ci {
								v = strings.ToLower(v)
							}
							if v == key {
								var resp any
								if field := e.v.Field(i); field.CanInterface() {
									resp = field.Interface()
								}
								return resp, true
							}
						}
					}
				}
			}
		}
	}
	return nil, false
}

// OneofIsInlineStructFieldFunc is a function for query.CustomIsInlineStructFieldFunc option to enable extracting values even if the oneof field name is omitted.
func OneofIsInlineStructFieldFunc() func(reflect.StructField) bool {
	return func(f reflect.StructField) bool {
		return f.Tag.Get("protobuf_oneof") != ""
	}
}
