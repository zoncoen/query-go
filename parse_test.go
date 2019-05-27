package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string]struct {
			src      string
			expected *Query
		}{
			"empty": {
				src:      "",
				expected: New(),
			},
			"key[index][index].key": {
				src:      "a[0][1].b",
				expected: New().Key("a").Index(0).Index(1).Key("b"),
			},
		}
		opt := cmp.AllowUnexported(Query{}, Key{}, Index{})
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				got, err := ParseString(test.src)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if diff := cmp.Diff(test.expected, got, opt); diff != "" {
					t.Errorf("differs: (-want +got)\n%s", diff)
				}
			})
		}
	})
}
