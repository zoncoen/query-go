package query

import "reflect"

func elem(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Ptr {
		switch v.Type().Elem().Kind() {
		case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
			v = v.Elem()
		}
	}
	return v
}
