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

func TestKey_Extract(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := map[string]struct {
			key    string
			v      interface{}
			expect interface{}
		}{
			"map[string]string": {
				key: "key",
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
				e := &Key{key: test.key}
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
			key string
			v   interface{}
		}{
			"target is nil": {
				key: "key",
				v:   nil,
			},
			"key not found": {
				key: "key",
				v:   map[string]string{},
			},
			"field not found": {
				key: "Invalid",
				v:   http.Request{},
			},
			"key extractor returns false": {
				key: "key",
				v:   &keyExtractor{},
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				e := &Key{key: test.key}
				v, ok := e.Extract(reflect.ValueOf(test.v))
				if ok {
					t.Fatalf("unexpected value: %#v", v)
				}
			})
		}
	})
}
