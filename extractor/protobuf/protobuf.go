package protobuf

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/zoncoen/query-go/v2"
)

// ExtractFunc is a function for query.CustomExtractFunc option to extract values by protobuf struct tag.
func ExtractFunc() func(query.ExtractFunc) query.ExtractFunc {
	return func(f query.ExtractFunc) query.ExtractFunc {
		return func(ctx context.Context, in reflect.Value) (reflect.Value, error) {
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
						v, err := f(ctx, reflect.ValueOf(&keyExtractor{v}))
						if err == nil {
							return v, nil
						}
						if !errors.Is(err, query.ErrNotFound) {
							// A failure is not an absence: do not fall back to
							// the next field or the plain struct lookup, which
							// would mask e.g. a canceled blocking extractor as
							// "not found".
							return reflect.Value{}, err
						}
					}
				}
			}
			return f(ctx, in)
		}
	}
}

type keyExtractor struct {
	v reflect.Value
}

// ExtractByKey implements the query.KeyExtractor interface.
func (e *keyExtractor) ExtractByKey(ctx context.Context, key string) (any, error) {
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
						if k == "name" || k == "json" {
							if ci {
								v = strings.ToLower(v)
							}
							if v == key {
								var resp any
								if field := e.v.Field(i); field.CanInterface() {
									resp = field.Interface()
								}
								return resp, nil
							}
						}
					}
				}
			}
		}
	}
	return nil, query.ErrNotFound
}

// OneofIsInlineStructFieldFunc is a function for query.CustomIsInlineStructFieldFunc option to enable extracting values even if the oneof field name is omitted.
func OneofIsInlineStructFieldFunc() func(reflect.StructField) bool {
	return func(f reflect.StructField) bool {
		return f.Tag.Get("protobuf_oneof") != ""
	}
}
