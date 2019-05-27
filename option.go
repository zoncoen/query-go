package query

import "reflect"

// Option represents an option for Query.
type Option func(*Query)

// CustomStructFieldNameGetter returns the Option to set f as custom function which gets struct field name.
// f is called by Key.Extract to get struct field name, if the target value is a struct.
func CustomStructFieldNameGetter(f func(f reflect.StructField) string) Option {
	return func(q *Query) {
		q.customStructFieldNameGetter = f
	}
}
