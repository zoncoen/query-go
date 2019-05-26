package parser

import (
	"strings"
	"testing"

	"github.com/zoncoen/query-go/token"
)

func TestScanner_Read(t *testing.T) {
	tests := map[string]struct {
		s         string
		buf       []rune
		expect    rune
		expectPos int
	}{
		"EOF": {
			expect:    eof,
			expectPos: 1,
		},
		"read": {
			s:         "abc",
			expect:    'a',
			expectPos: 2,
		},
		"read from buffer": {
			s:         "abc",
			buf:       []rune{'A', 'B', 'C'},
			expect:    'A',
			expectPos: 2,
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			s := newScanner(strings.NewReader(test.s))
			s.buf = test.buf
			if got := s.read(); got != test.expect {
				t.Errorf("expected %q but got %q", test.expect, got)
			}
			if got := s.pos; got != test.expectPos {
				t.Errorf("expected %d but got %d", test.expectPos, got)
			}
		})
	}
}

func TestScanner_Unread(t *testing.T) {
	s := &scanner{
		pos: 4,
	}
	s.unread('a')
	s.unread('b')
	s.unread('c')
	if got, expect := string(s.buf), "abc"; got != expect {
		t.Errorf("expected %q but got %q", expect, got)
	}
	if got, expect := s.pos, 1; got != expect {
		t.Errorf("expected %d but got %d", expect, got)
	}
}

func TestScanner_Scan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type result struct {
			pos int
			tok token.Token
			lit string
		}
		tests := map[string]struct {
			src      string
			expected []result
		}{
			"empty": {
				src:      "",
				expected: []result{},
			},
			"STRING": {
				src: `test`,
				expected: []result{
					{
						pos: 1,
						tok: token.STRING,
						lit: "test",
					},
				},
			},
			"STRING of integer": {
				src: `1`,
				expected: []result{
					{
						pos: 1,
						tok: token.STRING,
						lit: "1",
					},
				},
			},
			"[INT]": {
				src: `[10]`,
				expected: []result{
					{
						pos: 1,
						tok: token.LBRACK,
						lit: "[",
					},
					{
						pos: 2,
						tok: token.INT,
						lit: "10",
					},
					{
						pos: 4,
						tok: token.RBRACK,
						lit: "]",
					},
				},
			},
			"[STRING]": {
				src: `["test"]`,
				expected: []result{
					{
						pos: 1,
						tok: token.LBRACK,
						lit: "[",
					},
					{
						pos: 2,
						tok: token.STRING,
						lit: "test",
					},
					{
						pos: 8,
						tok: token.RBRACK,
						lit: "]",
					},
				},
			},
			"STRING[INT].STRING[STRING]": {
				src: `a[10].b["c"]`,
				expected: []result{
					{
						pos: 1,
						tok: token.STRING,
						lit: "a",
					},
					{
						pos: 2,
						tok: token.LBRACK,
						lit: "[",
					},
					{
						pos: 3,
						tok: token.INT,
						lit: "10",
					},
					{
						pos: 5,
						tok: token.RBRACK,
						lit: "]",
					},
					{
						pos: 6,
						tok: token.PERIOD,
						lit: ".",
					},
					{
						pos: 7,
						tok: token.STRING,
						lit: "b",
					},
					{
						pos: 8,
						tok: token.LBRACK,
						lit: "[",
					},
					{
						pos: 9,
						tok: token.STRING,
						lit: "c",
					},
					{
						pos: 12,
						tok: token.RBRACK,
						lit: "]",
					},
				},
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				s := newScanner(strings.NewReader(test.src))
				for i, e := range test.expected {
					pos, tok, lit := s.scan()
					if tok == token.EOF {
						t.Fatalf("[%d] unexpected EOF", i)
					}
					if got, expected := pos, e.pos; got != expected {
						t.Errorf(`[%d] expected position %d but got %d`, i, expected, got)
					}
					if got, expected := tok, e.tok; got != expected {
						t.Errorf(`[%d] expected token "%s" but got "%s"`, i, expected, got)
					}
					if got, expected := lit, e.lit; got != expected {
						t.Errorf(`[%d] expected literal "%s" but got "%s"`, i, expected, got)
					}
					if t.Failed() {
						t.FailNow()
					}
				}
				pos, tok, lit := s.scan()
				if tok != token.EOF {
					t.Fatalf(`expected EOF but got %d:%s:%s`, pos, tok, lit)
				}
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		tests := map[string]struct {
			src string
			pos int
			lit string
		}{
			"invalid integer index": {
				src: "[01]",
				pos: 2,
				lit: "01",
			},
			"invalid selector": {
				src: `[test]`,
				pos: 2,
				lit: "t",
			},
			"string not terminated": {
				src: `["test]`,
				pos: 8,
				lit: "",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				s := newScanner(strings.NewReader(test.src))
				for {
					pos, tok, lit := s.scan()
					if tok == token.EOF {
						t.Fatal("unexpected EOF")
					}
					if tok == token.ILLEGAL {
						if got, expected := pos, test.pos; got != expected {
							t.Errorf(`expected %d but got %d`, expected, got)
						}
						if got, expected := lit, test.lit; got != expected {
							t.Errorf(`expected "%s" but got "%s"`, expected, got)
						}
						break
					}
				}
			})
		}
	})
}
