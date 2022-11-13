package query

import (
	"reflect"
	"strings"
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
			target   interface{}
			expected interface{}
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
				query: New().Append(extractorFunc(func(v reflect.Value) (reflect.Value, bool) {
					return reflect.ValueOf((*int)(nil)), true
				})),
				expected: (*int)(nil),
			},
			"non-typed nil": {
				query: New().Append(extractorFunc(func(v reflect.Value) (reflect.Value, bool) {
					return reflect.ValueOf(nil), true
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
						return func(v reflect.Value) (reflect.Value, bool) {
							vv, ok := f(v)
							if ok {
								if vv.Kind() == reflect.String && vv.CanInterface() {
									return reflect.ValueOf("aaa" + vv.Interface().(string)), true
								}
							}
							return vv, true
						}
					}),
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(v reflect.Value) (reflect.Value, bool) {
							return reflect.ValueOf("bbb"), true
						}
					}),
				).Index(0),
				expected: "aaabbb",
			},
			"use CustomExtractFunc instead of CustomStructFieldNameGetter": {
				query: New(
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(v reflect.Value) (reflect.Value, bool) {
							v = elem(v)
							if v.Kind() == reflect.Struct {
								m := map[string]interface{}{}
								for i := 0; i < v.Type().NumField(); i++ {
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
								return f(reflect.ValueOf(m))
							}
							return f(v)
						}
					}),
				).Key("FOO_BAR"),
				target:   &testTags{FooBar: "aaa"},
				expected: "aaa",
			},
		}

		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				got, err := test.query.Extract(test.target)
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
			unexported struct{}
		}

		tests := map[string]struct {
			query  *Query
			target interface{}
		}{
			"unexported field (can not access)": {
				query: New().Append(extractorFunc(func(v reflect.Value) (reflect.Value, bool) {
					return reflect.ValueOf(test{}).FieldByName("unexported"), true
				})),
			},
			"CustomExtractFunc returns not ok": {
				query: New(
					CustomExtractFunc(func(f ExtractFunc) ExtractFunc {
						return func(v reflect.Value) (reflect.Value, bool) {
							vv, _ := f(v)
							return vv, false
						}
					}),
				).Index(0),
				target: []string{"a"},
			},
		}

		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				if _, err := test.query.Extract(test.target); err == nil {
					t.Fatal("no error")
				}
			})
		}
	})
}

type extractorFunc func(reflect.Value) (reflect.Value, bool)

func (f extractorFunc) Extract(v reflect.Value) (reflect.Value, bool) {
	return f(v)
}

func (f extractorFunc) String() string { return "" }
