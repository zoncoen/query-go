package query_test

import (
	"fmt"

	"github.com/zoncoen/query-go"
)

type orderedMap struct {
	elems []*elem
}

type elem struct {
	k, v interface{}
}

func (m *orderedMap) ExtractByKey(key string) (interface{}, bool) {
	for _, e := range m.elems {
		if k, ok := e.k.(string); ok {
			if k == key {
				return e.v, true
			}
		}
	}
	return nil, false
}

func ExampleKeyExtractor() {
	q := query.New().Key("key")
	v, _ := q.Extract(&orderedMap{
		elems: []*elem{{k: "key", v: "value"}},
	})
	fmt.Println(v)
	// Output:
	// value
}
