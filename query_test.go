package query

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestQuery_Append(t *testing.T) {
	q := New()
	q.extractors = make([]Extractor, 0, 1)
	q1 := q.Root().Key("1")
	q2 := q.Key("2")
	q3 := q1.Append(q2.Extractors()...).Append(&Key{key: "'3.0'"}, &Key{key: "4"})
	if got, expect := q.String(), ""; got != expect {
		t.Errorf(`expected "%s" but got "%s"`, expect, got)
	}
	if got, expect := q1.String(), "$.1"; got != expect {
		t.Errorf(`expected "%s" but got "%s"`, expect, got)
	}
	if got, expect := q2.String(), ".2"; got != expect {
		t.Errorf(`expected "%s" but got "%s"`, expect, got)
	}
	if got, expect := q3.String(), "$.1.2['\\'3.0\\''].4"; got != expect {
		t.Errorf(`expected "%s" but got "%s"`, expect, got)
	}
}

func TestQuery_Extract(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type debug struct {
			Prof map[string][]*keyExtractor
		}

		tests := map[string]struct {
			query    *Query
			target   any
			expected any
		}{
			"query is nil": {
				query:    nil,
				target:   "value",
				expected: "value",
			},
			"target is nil": {
				query:    New(),
				target:   nil,
				expected: nil,
			},
			"empty query": {
				query:    New(),
				target:   "value",
				expected: "value",
			},
			"typed nil": {
				query: New().Append(extractorFunc(func(_ context.Context, v reflect.Value) (reflect.Value, error) {
					return reflect.ValueOf((*int)(nil)), nil
				})),
				expected: (*int)(nil),
			},
			"non-typed nil": {
				query: New().Append(extractorFunc(func(_ context.Context, v reflect.Value) (reflect.Value, error) {
					return reflect.ValueOf(nil), nil
				})),
				expected: nil,
			},
			"complex": {
				query: New().Key("Prof").Key("heap").Index(1).Key("sum%"),
				target: &debug{
					Prof: map[string][]*keyExtractor{
						"heap": {
							{v: "80%"}, {v: "100%"},
						},
					},
				},
				expected: "100%",
			},
			"with $": {
				query:    New().Root().Key("foo"),
				target:   map[string]string{"foo": "aaa"},
				expected: "aaa",
			},
			"CaseInsensitive": {
				query:    New(CaseInsensitive()).Key("foo"),
				target:   map[string]string{"Foo": "aaa"},
				expected: "aaa",
			},
			"ExtractByStructTag": {
				query:    New(CaseInsensitive(), ExtractByStructTag("json")).Key("FOO_BAR"),
				target:   &testTags{FooBar: "aaa"},
				expected: "aaa",
			},
			"CustomExtractFunc": {
				query: New(
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
							vv, err := f(ctx, v)
							if err == nil {
								if vv.Kind() == reflect.String && vv.CanInterface() {
									return reflect.ValueOf("aaa" + vv.Interface().(string)), nil
								}
							}
							return vv, nil
						}
					}),
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
							return reflect.ValueOf("bbb"), nil
						}
					}),
				).Index(0),
				expected: "aaabbb",
			},
			"use CustomExtractFunc instead of CustomStructFieldNameGetter": {
				query: New(
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
							v = elem(v)
							if v.Kind() == reflect.Struct {
								m := map[string]any{}
								for i := range v.Type().NumField() {
									field := v.Type().FieldByIndex([]int{i})
									if s := field.Tag.Get("json"); s != "" {
										name, _, _ := strings.Cut(s, ",")
										if name != "" {
											f := v.FieldByIndex([]int{i})
											if f.CanInterface() {
												m[strings.ToUpper(name)] = f.Interface()
											}
										}
									}
								}
								return f(ctx, reflect.ValueOf(m))
							}
							return f(ctx, v)
						}
					}),
				).Key("FOO_BAR"),
				target:   &testTags{FooBar: "aaa"},
				expected: "aaa",
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				got, err := test.query.Extract(context.Background(), test.target)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if diff := cmp.Diff(test.expected, got); diff != "" {
					t.Errorf("differs: (-want +got)\n%s", diff)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		type test struct {
			unexported struct{} //nolint:unused // never read; the failure cases exercise access to an unexported field
		}

		tests := map[string]struct {
			query  *Query
			target any
		}{
			"unexported field (can not access)": {
				query: New().Append(extractorFunc(func(_ context.Context, v reflect.Value) (reflect.Value, error) {
					return reflect.ValueOf(test{}).FieldByName("unexported"), nil
				})),
			},
			"CustomExtractFunc returns ErrNotFound": {
				query: New(
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
							vv, _ := f(ctx, v)
							return vv, ErrNotFound
						}
					}),
				).Index(0),
				target: []string{"a"},
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				if _, err := test.query.Extract(context.Background(), test.target); err == nil {
					t.Fatal("no error")
				}
			})
		}
	})
}

type extractorFunc func(context.Context, reflect.Value) (reflect.Value, error)

func (f extractorFunc) Extract(ctx context.Context, v reflect.Value) (reflect.Value, error) {
	return f(ctx, v)
}

func (f extractorFunc) String() string { return "" }

type ctxKeyTest struct{}

// ctxValueExtractor returns the value stored in the context, to verify the
// caller's context reaches the extractor.
type ctxValueExtractor struct{}

func (e *ctxValueExtractor) ExtractByIndex(ctx context.Context, _ int) (any, error) {
	if v, ok := ctx.Value(ctxKeyTest{}).(string); ok {
		return v, nil
	}
	return nil, ErrNotFound
}

func TestQuery_Extract_ContextPropagation(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKeyTest{}, "from-caller")
	got, err := New().Index(0).Extract(ctx, &ctxValueExtractor{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != "from-caller" {
		t.Fatalf("expected propagated context value but got %v", got)
	}
}

func TestQuery_Extract_NotFound(t *testing.T) {
	tests := map[string]struct {
		query          *Query
		target         any
		expectFailedAt string
	}{
		"fails at the last extractor": {
			query:          New().Key("a").Key("b"),
			target:         map[string]any{"a": map[string]any{}},
			expectFailedAt: ".a.b",
		},
		"fails at an intermediate extractor": {
			query:          New().Key("a").Key("b"),
			target:         map[string]any{},
			expectFailedAt: ".a",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := test.query.Extract(context.Background(), test.target)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("expected ErrNotFound but got: %s", err)
			}
			var nfe *NotFoundError
			if !errors.As(err, &nfe) {
				t.Fatalf("expected *NotFoundError but got %T", err)
			}
			if got, expect := nfe.Query, test.query.String(); got != expect {
				t.Errorf("Query: expected %q but got %q", expect, got)
			}
			if got := nfe.FailedAt; got != test.expectFailedAt {
				t.Errorf("FailedAt: expected %q but got %q", test.expectFailedAt, got)
			}
			if got, expect := err.Error(), fmt.Sprintf("%q not found", test.query.String()); got != expect {
				t.Errorf("message: expected %q but got %q", expect, got)
			}
		})
	}
}

// interruptedExtractor simulates a blocking extractor whose wait was cut
// short by the context ending.
type interruptedExtractor struct{}

func (e *interruptedExtractor) ExtractByIndex(ctx context.Context, _ int) (any, error) {
	return nil, ctx.Err()
}

func TestQuery_Extract_FailurePropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New().Key("messages").Index(0).Extract(ctx, map[string]any{
		"messages": &interruptedExtractor{},
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("a failure must not match ErrNotFound: %s", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected the context error to propagate but got: %s", err)
	}
	if got, expect := err.Error(), ".messages[0]: context canceled"; got != expect {
		t.Errorf("expected %q but got %q", expect, got)
	}
}

func TestQuery_Extract_Concurrent(t *testing.T) {
	q := New().Key("a").Index(1)
	target := map[string][]string{"a": {"x", "y"}}
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				got, err := q.Extract(context.Background(), target)
				if err != nil || got != "y" {
					t.Errorf("unexpected result: %v, %v", got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// nestedExtractor resolves its key against a nested map with a sub-query
// configured from the caller's options.
type nestedExtractor struct {
	v any
}

func (e *nestedExtractor) ExtractByKey(ctx context.Context, key string) (any, error) {
	return New(OptionsFromContext(ctx)...).Key(key).Extract(ctx, e.v)
}

func TestOptionsFromContext(t *testing.T) {
	target := map[string]any{
		"nested": &nestedExtractor{v: &testTags{FooBar: "aaa"}},
	}
	got, err := New(ExtractByStructTag("json")).Key("nested").Key("foo_bar").Extract(context.Background(), target)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != "aaa" {
		t.Fatalf("expected the caller's struct-tag option to reach the sub-query, got %v", got)
	}
	if opts := OptionsFromContext(context.Background()); opts != nil {
		t.Fatalf("expected nil outside an extraction, got %v", opts)
	}
}
