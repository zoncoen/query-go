package query

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type keyExtractor struct {
	v interface{}
}

func (f *keyExtractor) ExtractByKey(_ string) (interface{}, bool) {
	if f.v != nil {
		return f.v, true
	}
	return nil, false
}

type keyExtractorContext struct {
	v map[string]any
}

func (f *keyExtractorContext) ExtractByKey(ctx context.Context, name string) (interface{}, bool) {
	if f.v != nil {
		if v, ok := f.v[name]; ok {
			return v, true
		}
		if IsCaseInsensitive(ctx) {
			name = strings.ToLower(name)
			for k, v := range f.v {
				if strings.ToLower(k) == name {
					return v, true
				}
			}
		}
	}
	return nil, false
}

type testTags struct {
	FooBar string `json:"foo_bar" yaml:"fooBar,omitempty"`
	AnonymousField
	M      map[string]string `json:",inline"`
	Inline map[string]string

	state struct{}
	State string `json:"state"`
}

type AnonymousField struct {
	S string
}

func TestKey_Extract(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := map[string]struct {
			key                string
			caseInsensitive    bool
			structTags         []string
			customExtractFuncs []func(ExtractFunc) ExtractFunc
			isInlineFuncs      []func(reflect.StructField) bool
			v                  interface{}
			expect             interface{}
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
				v: map[interface{}]interface{}{
					0:     0,
					"key": 1,
				},
				expect: 1,
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
						return func(v reflect.Value) (reflect.Value, bool) {
							if v.CanInterface() {
								if vv, ok := v.Interface().(map[string]string); ok {
									mp := map[string]any{}
									for k, v := range vv {
										mp[k] = v + v
									}
									if v, ok := f(reflect.ValueOf(&keyExtractorContext{v: mp})); ok {
										return v, true
									}
								}
							}
							return f(v)
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
				v: &keyExtractorContext{
					v: map[string]any{
						"KEY": "value",
					},
				},
				expect: "value",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Key{
					key:                test.key,
					caseInsensitive:    test.caseInsensitive,
					structTags:         test.structTags,
					customExtractFuncs: test.customExtractFuncs,
					isInlineFuncs:      test.isInlineFuncs,
				}
				v, ok := e.Extract(reflect.ValueOf(test.v))
				if !ok {
					t.Fatal("not found")
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
			v             interface{}
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
				v: &keyExtractorContext{
					v: map[string]any{
						"KEY": "value",
					},
				},
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Key{
					key:        test.key,
					structTags: test.structTags,
				}
				v, ok := e.Extract(reflect.ValueOf(test.v))
				if ok {
					t.Fatalf("unexpected value: %#v", v)
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
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			k := &Key{key: test.key}
			if got := k.String(); got != test.expect {
				t.Errorf("expect %q but got %q", test.expect, got)
			}
		})
	}
}
