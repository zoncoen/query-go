package query

import "context"

var caseInsensitiveKey = struct{}{}

func withCaseInsensitive(ctx context.Context, b bool) context.Context {
	return context.WithValue(ctx, caseInsensitiveKey, b)
}

// IsCaseInsensitive reports whether case-insensitive querying is enabled or not.
func IsCaseInsensitive(ctx context.Context) bool {
	if b, ok := ctx.Value(caseInsensitiveKey).(bool); ok {
		return b
	}
	return false
}
