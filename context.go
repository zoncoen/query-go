package query

import "context"

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
