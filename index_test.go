package query

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type indexExtractor struct {
	v interface{}
}

func (f *indexExtractor) ExtractByIndex(_ int) (interface{}, bool) {
	if f.v != nil {
		return f.v, true
	}
	return nil, false
}

func TestIndex_Extract(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := map[string]struct {
			index  int
			v      interface{}
			expect interface{}
		}{
			"slice": {
				index: 0,
				v: []int{
					0, 1, 2,
				},
				expect: 0,
			},
			"array": {
				index: 1,
				v: [3]int{
					0, 1, 2,
				},
				expect: 1,
			},
			"array pointer": {
				index: 2,
				v: &[3]int{
					0, 1, 2,
				},
				expect: 2,
			},
			"index extractor": {
				index:  10,
				v:      &indexExtractor{v: "value"},
				expect: "value",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Index{index: test.index}
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
			index int
			v     interface{}
		}{
			"target is nil": {
				index: 0,
				v:     nil,
			},
			"slice has not index": {
				index: 0,
				v:     []int{},
			},
			"array has not index": {
				index: 1,
				v:     [1]int{0},
			},
			"index extractor returns false": {
				index: 10,
				v:     &indexExtractor{},
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Index{index: test.index}
				v, ok := e.Extract(reflect.ValueOf(test.v))
				if ok {
					t.Fatalf("unexpected value: %#v", v)
				}
			})
		}
	})
}
