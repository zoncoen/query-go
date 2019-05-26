package parser

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zoncoen/query-go/ast"
)

func TestParser_Parse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string]struct {
			src      string
			expected ast.Node
		}{
			"empty": {
				src:      "",
				expected: nil,
			},
			"a selector": {
				src: ".selector",
				expected: &ast.Selector{
					ValuePos: 1,
					Sel:      "selector",
				},
			},
			"a selector w/o period": {
				src: "selector",
				expected: &ast.Selector{
					ValuePos: 1,
					Sel:      "selector",
				},
			},
			"an index": {
				src: "[0]",
				expected: &ast.Index{
					ValuePos: 1,
					Index:    0,
				},
			},
			"selectors": {
				src: "a.b.c",
				expected: &ast.Selector{
					ValuePos: 4,
					X: &ast.Selector{
						ValuePos: 2,
						X: &ast.Selector{
							ValuePos: 1,
							Sel:      "a",
						},
						Sel: "b",
					},
					Sel: "c",
				},
			},
			"bracket selectors": {
				src: `["0"]["1"]["2"]`,
				expected: &ast.Selector{
					ValuePos: 11,
					X: &ast.Selector{
						ValuePos: 6,
						X: &ast.Selector{
							ValuePos: 1,
							Sel:      "0",
						},
						Sel: "1",
					},
					Sel: "2",
				},
			},
			"indices": {
				src: "[0][1][2]",
				expected: &ast.Index{
					ValuePos: 7,
					X: &ast.Index{
						ValuePos: 4,
						X: &ast.Index{
							ValuePos: 1,
							Index:    0,
						},
						Index: 1,
					},
					Index: 2,
				},
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				p := NewParser(strings.NewReader(test.src))
				got, err := p.Parse()
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if diff := cmp.Diff(test.expected, got); diff != "" {
					t.Errorf("result differs: (-want +got)\n%s", diff)
				}
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		tests := map[string]struct {
			src string
			pos int
		}{
			"expected ] but got EOF": {
				src: "[0",
				pos: 3,
			},
			"expected ] but got [": {
				src: "[0[1]",
				pos: 3,
			},
			`expected " but got EOF`: {
				src: `["0`,
				pos: 4,
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				p := NewParser(strings.NewReader(test.src))
				_, err := p.Parse()
				if err == nil {
					t.Fatal("expected error")
				}
				errs, ok := err.(Errors)
				if !ok {
					t.Fatalf("expected parse errors: %s", err)
				}
				if got, expected := errs[0].pos, test.pos; got != expected {
					t.Fatalf("expected %d but got %d: %s", expected, got, err)
				}
			})
		}
	})
}
