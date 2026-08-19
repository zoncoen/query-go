package query

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type indexExtractor struct {
	v interface{}
}

type namedIntKey int

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
			"negative index accesses the slice from the end": {
				index: -1,
				v: []int{
					0, 1, 2,
				},
				expect: 2,
			},
			"negative index accesses the array from the end": {
				index: -3,
				v: [3]int{
					0, 1, 2,
				},
				expect: 0,
			},
			"integer-keyed map": {
				index: 1,
				v: map[int]string{
					1: "one",
				},
				expect: "one",
			},
			"integer-keyed map with a negative literal key": {
				index: -1,
				v: map[int64]string{
					-1: "minus one",
				},
				expect: "minus one",
			},
			"map with a named integer key type": {
				index: 1,
				v: map[namedIntKey]string{
					1: "one",
				},
				expect: "one",
			},
			"unsigned-keyed map": {
				index: 200,
				v: map[uint8]string{
					200: "value",
				},
				expect: "value",
			},
			"interface-keyed map holding an int key": {
				index: 1,
				v: map[interface{}]string{
					"1": "string",
					1:   "int",
				},
				expect: "int",
			},
			"index extractor": {
				index:  10,
				v:      &indexExtractor{v: "value"},
				expect: "value",
			},
			"index extractor receives a negative index as given": {
				index:  -1,
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
			"negative index out of range": {
				index: -4,
				v:     []int{0, 1, 2},
			},
			"negative index into empty slice": {
				index: -1,
				v:     []int{},
			},
			"map key not found": {
				index: 2,
				v:     map[int]string{1: "one"},
			},
			"string-keyed map is not indexable": {
				index: 0,
				v:     map[string]string{"0": "value"},
			},
			"float-keyed map is not indexable": {
				index: 1,
				v:     map[float64]string{1: "value"},
			},
			"index overflows the map key type": {
				index: 300,
				v:     map[int8]string{44: "truncated 300"},
			},
			"negative index into an unsigned-keyed map": {
				index: -1,
				v:     map[uint]string{0: "value"},
			},
			"interface-keyed map holding a differently typed integer key": {
				index: 1,
				v:     map[interface{}]string{int8(1): "int8"},
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
