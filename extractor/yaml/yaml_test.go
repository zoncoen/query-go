package yaml

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/zoncoen/query-go"
)

func TestMapSliceExtractFunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string]struct {
			query  *query.Query
			v      interface{}
			expect interface{}
		}{
			"yaml.MapSlice": {
				query: query.New(
					query.CustomExtractFunc(MapSliceExtractFunc(false)),
				).Key("foo"),
				v: yaml.MapSlice{
					yaml.MapItem{
						Key:   "foo",
						Value: "aaa",
					},
				},
				expect: "aaa",
			},
			"yaml.MapSlice (case-insensitive)": {
				query: query.New(
					query.CustomExtractFunc(MapSliceExtractFunc(true)),
				).Key("foo"),
				v: yaml.MapSlice{
					yaml.MapItem{
						Key:   "Foo",
						Value: "aaa",
					},
				},
				expect: "aaa",
			},
			"not yaml.MapSlice": {
				query: query.New(
					query.CustomExtractFunc(MapSliceExtractFunc(false)),
				).Index(1),
				v:      []int{1, 2},
				expect: 2,
			},
			"not slice": {
				query: query.New(
					query.CaseInsensitive(),
					query.CustomExtractFunc(MapSliceExtractFunc(false)),
				).Key("foo"),
				v: map[string]string{
					"foo": "aaa",
				},
				expect: "aaa",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				got, err := test.query.Extract(test.v)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if got != test.expect {
					t.Errorf("expect %v but got %v", test.expect, got)
				}
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		tests := map[string]struct {
			query  *query.Query
			v      interface{}
			expect string
		}{
			"yaml.MapSlice": {
				query: query.New(
					query.CustomExtractFunc(MapSliceExtractFunc(false)),
				).Key("foo"),
				v: yaml.MapSlice{
					yaml.MapItem{
						Key:   "Foo",
						Value: "aaa",
					},
				},
				expect: `".foo" not found`,
			},
			"not yaml.MapSlice": {
				query: query.New(
					query.CustomExtractFunc(MapSliceExtractFunc(false)),
				).Key("bar"),
				v:      []int{},
				expect: `".bar" not found`,
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				_, err := test.query.Extract(test.v)
				if err == nil {
					t.Fatal("no error")
				}
				if got := err.Error(); !strings.Contains(got, test.expect) {
					t.Errorf("expect %v but got %v", test.expect, got)
				}
			})
		}
	})
}
