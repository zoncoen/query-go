package query_test

import (
	"context"
	"fmt"

	"github.com/zoncoen/query-go/v2"
)

type orderedMap struct {
	elems []*elem
}

type elem struct {
	k, v any
}

func (m *orderedMap) ExtractByKey(_ context.Context, key string) (any, error) {
	for _, e := range m.elems {
		if k, ok := e.k.(string); ok {
			if k == key {
				return e.v, nil
			}
		}
	}
	return nil, query.ErrNotFound
}

func ExampleKeyExtractor() {
	q := query.New().Key("key")
	v, _ := q.Extract(context.Background(), &orderedMap{
		elems: []*elem{{k: "key", v: "value"}},
	})
	fmt.Println(v)
	// Output:
	// value
}
