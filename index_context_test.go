package query

import (
	"context"
	"testing"
)

type ctxKey struct{}

// indexExtractorContext records the context it receives and returns a value
// derived from it, to verify the caller's context is propagated.
type indexExtractorContext struct {
	gotCtx context.Context
}

func (e *indexExtractorContext) ExtractByIndex(ctx context.Context, _ int) (any, bool) {
	e.gotCtx = ctx
	v, _ := ctx.Value(ctxKey{}).(string)
	return v, true
}

func TestIndex_ExtractContext(t *testing.T) {
	e := &indexExtractorContext{}
	q := New().Index(0)

	ctx := context.WithValue(context.Background(), ctxKey{}, "from-caller")
	got, err := q.ExtractContext(ctx, e)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != "from-caller" {
		t.Fatalf("expected propagated context value but got %v", got)
	}
	if e.gotCtx == nil {
		t.Fatal("expected the extractor to receive a context")
	}
}

func TestIndex_Extract_FallsBackToBackground(t *testing.T) {
	e := &indexExtractorContext{}
	q := New().Index(0)

	// The context-less Extract must still reach IndexExtractorContext with a
	// background context, preserving behavior for callers that don't pass one.
	if _, err := q.Extract(e); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if e.gotCtx == nil {
		t.Fatal("expected the extractor to receive a background context")
	}
	if e.gotCtx.Value(ctxKey{}) != nil {
		t.Fatal("expected a background context with no caller values")
	}
}

// plainIndexExtractor only implements the context-less interface.
type plainIndexExtractor struct{ called bool }

func (e *plainIndexExtractor) ExtractByIndex(_ int) (interface{}, bool) {
	e.called = true
	return "plain", true
}

func TestIndex_ExtractContext_FallsBackToPlainExtractor(t *testing.T) {
	e := &plainIndexExtractor{}
	q := New().Index(0)
	got, err := q.ExtractContext(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !e.called || got != "plain" {
		t.Fatalf("expected plain IndexExtractor to be used, got %v (called=%v)", got, e.called)
	}
}
