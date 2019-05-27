package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestQuery_Extract(t *testing.T) {
	type debug struct {
		Prof map[string][]*keyExtractor
	}

	tests := map[string]struct {
		query    *Query
		target   interface{}
		expected interface{}
	}{
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
}
