package query_test

import (
	"fmt"

	"github.com/zoncoen/query-go/v2"
)

type S struct {
	Maps []map[string]map[string]string
}

func ExampleParseString() {
	q, err := query.ParseString(`$.Maps[0].key['.key\'']`)
	if err == nil {
		v, _ := q.Extract(&S{
			Maps: []map[string]map[string]string{
				{"key": map[string]string{
					".key'": "value",
				}},
			},
		})
		fmt.Println(v)
		// Output:
		// value
	}
}
