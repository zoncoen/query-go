package query

import "reflect"

// Option represents an option for Query.
type Option func(*Query)

// CaseInsensitive returns the Option to match case insensitivity.
func CaseInsensitive() Option {
	return func(q *Query) {
		q.caseInsensitive = true
	}
}

// ExtractByStructTag returns the Option to allow extracting by struct tag.
func ExtractByStructTag(tagNames ...string) Option {
	return func(q *Query) {
		q.structTags = append(q.structTags, tagNames...)
	}
}

// CustomExtractFunc returns the Option to customize the behavior of extractors.
func CustomExtractFunc(f func(ExtractFunc) ExtractFunc) Option {
	return func(q *Query) {
		q.customExtractFuncs = append(q.customExtractFuncs, f)
	}
}

// CustomStructFieldNameGetter returns the Option to set f as custom function which gets struct field name.
// f is called by Key.Extract to get struct field name, if the target value is a struct.
//
// Deprecated: Use CustomExtractFunc instead.
func CustomStructFieldNameGetter(f func(f reflect.StructField) string) Option {
	return func(q *Query) {
		q.customStructFieldNameGetter = f
	}
}

// CustomIsInlineStructFieldFunc returns the Option to customize the behavior of extractors.
func CustomIsInlineStructFieldFunc(f func(reflect.StructField) bool) Option {
	return func(q *Query) {
		q.customIsInlineFuncs = append(q.customIsInlineFuncs, f)
	}
}
