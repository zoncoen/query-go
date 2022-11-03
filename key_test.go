package query

import (
	"net/http"
	"reflect"
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

type testTags struct {
	FooBar string `json:"foo_bar" yaml:"fooBar,omitempty"`
	AnonymousField
	M map[string]string `json:",inline"`
}

type AnonymousField struct {
	S string
}

func TestKey_Extract(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := map[string]struct {
			key             string
			caseInsensitive bool
			structTags      []string
			v               interface{}
			expect          interface{}
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
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Key{
					key:             test.key,
					caseInsensitive: test.caseInsensitive,
					structTags:      test.structTags,
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
			key        string
			structTags []string
			v          interface{}
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
