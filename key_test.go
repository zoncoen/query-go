package query

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type keyExtractor struct {
	v any
}

func (f *keyExtractor) ExtractByKey(_ context.Context, _ string) (any, error) {
	if f.v != nil {
		return f.v, nil
	}
	return nil, ErrNotFound
}

type caseInsensitiveKeyExtractor struct {
	v map[string]any
}

func (f *caseInsensitiveKeyExtractor) ExtractByKey(ctx context.Context, name string) (any, error) {
	if f.v != nil {
		if v, ok := f.v[name]; ok {
			return v, nil
		}
		if IsCaseInsensitive(ctx) {
			name = strings.ToLower(name)
			for k, v := range f.v {
				if strings.ToLower(k) == name {
					return v, nil
				}
			}
		}
	}
	return nil, ErrNotFound
}

type testTags struct {
	FooBar string `json:"foo_bar" yaml:"fooBar,omitempty"`
	AnonymousField
	M      map[string]string `json:",inline"`
	Inline map[string]string

	state struct{} //nolint:unused // never read; its unexported field name is what the fallthrough test matches against
	State string   `json:"state"`
}

type AnonymousField struct {
	S string
}

type namedKey string

func TestKey_Extract(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := map[string]struct {
			key                string
			caseInsensitive    bool
			structTags         []string
			customExtractFuncs []func(ExtractFunc) ExtractFunc
			isInlineFuncs      []func(reflect.StructField) bool
			v                  any
			expect             any
		}{
			"map[string]string": {
				key: "key",
				v: map[string]string{
					"key": "value",
				},
				expect: "value",
			},
			"map[string]string (case-insensitive)": {
				key:             "KEY",
				caseInsensitive: true,
				v: map[string]string{
					"key": "value",
				},
				expect: "value",
			},
			"map[interface{}]interface{}": {
				key: "key",
				v: map[any]any{
					0:     0,
					"key": 1,
				},
				expect: 1,
			},
			"map with a named string key type": {
				key: "key",
				v: map[namedKey]string{
					"key": "value",
				},
				expect: "value",
			},
			"map[string]string (case-insensitive, exact match wins)": {
				key:             "key",
				caseInsensitive: true,
				v: map[string]string{
					"key": "exact",
					"KEY": "upper",
				},
				expect: "exact",
			},
			"map[string]string (case-insensitive, smallest key wins)": {
				key:             "key",
				caseInsensitive: true,
				v: map[string]string{
					"KEY": "upper",
					"Key": "title",
				},
				expect: "upper",
			},
			"struct": {
				key:    "Method",
				v:      http.Request{Method: http.MethodGet},
				expect: http.MethodGet,
			},
			"struct (case-insensitive)": {
				key:             "method",
				caseInsensitive: true,
				v:               http.Request{Method: http.MethodGet},
				expect:          http.MethodGet,
			},
			"struct (anonymous field)": {
				key:             "AnonymousField",
				caseInsensitive: true,
				v: testTags{
					AnonymousField: AnonymousField{
						S: "aaa",
					},
				},
				expect: AnonymousField{
					S: "aaa",
				},
			},
			"struct (anonymous field's field)": {
				key:             "S",
				caseInsensitive: true,
				v: testTags{
					AnonymousField: AnonymousField{
						S: "aaa",
					},
				},
				expect: "aaa",
			},
			"struct (strcut tag)": {
				key:        "foo_bar",
				structTags: []string{"json", "yaml"},
				v: testTags{
					FooBar: "xxx",
				},
				expect: "xxx",
			},
			"struct (strcut tag with option)": {
				key:        "fooBar",
				structTags: []string{"json", "yaml"},
				v: testTags{
					FooBar: "xxx",
				},
				expect: "xxx",
			},
			"struct (inline strcut tag option)": {
				key:        "aaa",
				structTags: []string{"json", "yaml"},
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
				},
				expect: "xxx",
			},
			"struct (inline with custom extract funcs)": {
				key:        "aaa",
				structTags: []string{"json", "yaml"},
				customExtractFuncs: []func(ExtractFunc) ExtractFunc{
					func(f ExtractFunc) ExtractFunc {
						return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
							if v.CanInterface() {
								if vv, ok := v.Interface().(map[string]string); ok {
									mp := map[string]any{}
									for k, v := range vv {
										mp[k] = v + v
									}
									if v, err := f(ctx, reflect.ValueOf(&caseInsensitiveKeyExtractor{v: mp})); err == nil {
										return v, nil
									}
								}
							}
							return f(ctx, v)
						}
					},
				},
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
				},
				expect: "xxxxxx",
			},
			"struct (custom inline func)": {
				key: "aaa",
				isInlineFuncs: []func(reflect.StructField) bool{
					func(f reflect.StructField) bool {
						return f.Name == "Inline"
					},
				},
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
					Inline: map[string]string{
						"aaa": "yyy",
					},
				},
				expect: "yyy",
			},
			"struct (fallthrough unexported field)": {
				key:        "state",
				structTags: []string{"json"},
				v: testTags{
					State: "ready",
				},
				expect: "ready",
			},
			"struct pointer": {
				key:    "Method",
				v:      &http.Request{Method: http.MethodGet},
				expect: http.MethodGet,
			},
			"key extractor": {
				key:    "key",
				v:      &keyExtractor{v: "value"},
				expect: "value",
			},
			"key extractor context": {
				key:             "key",
				caseInsensitive: true,
				v: &caseInsensitiveKeyExtractor{
					v: map[string]any{
						"KEY": "value",
					},
				},
				expect: "value",
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				e := &Key{
					key:                test.key,
					caseInsensitive:    test.caseInsensitive,
					structTags:         test.structTags,
					customExtractFuncs: test.customExtractFuncs,
					isInlineFuncs:      test.isInlineFuncs,
				}
				v, err := e.Extract(context.Background(), reflect.ValueOf(test.v))
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if diff := cmp.Diff(test.expect, v.Interface()); diff != "" {
					t.Errorf("differs: (-want +got)\n%s", diff)
				}
			})
		}
	})
	t.Run("not found", func(t *testing.T) {
		tests := map[string]struct {
			key           string
			structTags    []string
			isInlineFuncs []func(reflect.StructField) bool
			v             any
		}{
			"target is nil": {
				key: "key",
				v:   nil,
			},
			"key not found": {
				key: "key",
				v: map[string]string{
					"Key": "case sensitive",
				},
			},
			"integer-keyed map is not accessible by a string key": {
				key: "1",
				v: map[int]string{
					1: "value",
				},
			},
			"placeholder string must not match a non-string key": {
				key: "<int Value>",
				v: map[int]string{
					1: "value",
				},
			},
			"field not found": {
				key: "Invalid",
				v:   http.Request{},
			},
			"key extractor returns false": {
				key: "key",
				v:   &keyExtractor{},
			},
			"strcut tag option": {
				key:        "FOO_BAR",
				structTags: []string{"json", "yaml"},
				v: testTags{
					FooBar: "xxx",
				},
			},
			"struct (anonymous field's field)": {
				key: "s",
				v: testTags{
					AnonymousField: AnonymousField{
						S: "aaa",
					},
				},
			},
			"inline": {
				key:        "AAA",
				structTags: []string{"json", "yaml"},
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
				},
			},
			"inline (not contains json tag)": {
				key:        "aaa",
				structTags: []string{"yaml"},
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
				},
			},
			"inline (no custom inline func)": {
				key: "aaa",
				v: testTags{
					M: map[string]string{
						"aaa": "xxx",
					},
					Inline: map[string]string{
						"aaa": "yyy",
					},
				},
			},
			"key extractor context (case sensitive)": {
				key: "key",
				v: &caseInsensitiveKeyExtractor{
					v: map[string]any{
						"KEY": "value",
					},
				},
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				e := &Key{
					key:        test.key,
					structTags: test.structTags,
				}
				v, err := e.Extract(context.Background(), reflect.ValueOf(test.v))
				if err == nil {
					t.Fatalf("unexpected value: %#v", v)
				}
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("expected ErrNotFound but got: %s", err)
				}
			})
		}
	})
}

func TestKey_String(t *testing.T) {
	tests := map[string]struct {
		key    string
		expect string
	}{
		"simple": {
			key:    "aaa",
			expect: ".aaa",
		},
		"[": {
			key:    "[",
			expect: "['[']",
		},
		".": {
			key:    ".",
			expect: "['.']",
		},
		"\\": {
			key:    "\\",
			expect: "['\\\\']",
		},
		"'": {
			key:    "'",
			expect: "['\\'']",
		},
		"]": {
			key:    "]",
			expect: "[']']",
		},
		"$": {
			key:    "$foo",
			expect: "['$foo']",
		},
		"empty": {
			key:    "",
			expect: "['']",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := &Key{key: test.key}
			if got := k.String(); got != test.expect {
				t.Errorf("expect %q but got %q", test.expect, got)
			}
		})
	}
}

func TestKey_Extract_UnexportedEmbeddedStruct(t *testing.T) {
	type inner struct{ Foo string }
	v := struct{ inner }{inner{Foo: "value"}}
	// The inline expansion feeds the read-only value of the unexported
	// embedded field back into Key.Extract; it must not panic in Interface.
	// Exported fields promoted through an unexported embedded field are
	// interfaceable, so the extraction succeeds.
	got, err := New().Key("Foo").Extract(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expect := "value"; got != expect {
		t.Fatalf("expected %q but got %q", expect, got)
	}
}

func TestKey_String_RoundTrip(t *testing.T) {
	keys := []string{"aaa", "[", "]", ".", "\\", "'", "$", "$foo", "a$b", "a]b", "", "foo.bar-baz"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			q := New().Key(key)
			got, err := ParseString(q.String())
			if err != nil {
				t.Fatalf("failed to reparse %q: %s", q.String(), err)
			}
			if diff := cmp.Diff(q, got, cmp.AllowUnexported(Query{}, Key{}, Index{})); diff != "" {
				t.Errorf("%q does not round-trip: (-want +got)\n%s", q.String(), diff)
			}
		})
	}
}
