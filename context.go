package query

import (
	"context"
	"slices"
)

// caseInsensitiveKey is a named type so that the context key cannot collide
// with a key of another package: values of an anonymous struct{} type all
// compare equal regardless of where they are created.
type caseInsensitiveKey struct{}

func withCaseInsensitive(ctx context.Context, b bool) context.Context {
	return context.WithValue(ctx, caseInsensitiveKey{}, b)
}

// IsCaseInsensitive reports whether case-insensitive querying is enabled or not.
func IsCaseInsensitive(ctx context.Context) bool {
	if b, ok := ctx.Value(caseInsensitiveKey{}).(bool); ok {
		return b
	}
	return false
}

type optionsKey struct{}

func withOptions(ctx context.Context, opts []Option) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

// OptionsFromContext returns the options of the query that initiated the
// current extraction, or nil outside an extraction. An extractor
// implementation that builds a sub-query — e.g. a KeyExtractor delegating
// into a nested structure — can pass them to New so that the caller's
// configuration (struct tags, custom extract funcs, ...) applies to the
// nested extraction as well, without any global registry:
//
//	q := query.New(query.OptionsFromContext(ctx)...).Key("nested")
//
// The returned slice is a copy: appending to it or mutating it in place
// cannot affect other extractors in the same extraction.
func OptionsFromContext(ctx context.Context) []Option {
	if opts, ok := ctx.Value(optionsKey{}).([]Option); ok {
		return slices.Clone(opts)
	}
	return nil
}
