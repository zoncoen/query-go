package parser

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zoncoen/query-go/token"
)

// eof represents invalid code points.
var eof = unicode.ReplacementChar

type scanner struct {
	r              *bufio.Reader
	pos            int
	buf            []rune
	isReadingIndex bool
}

func newScanner(r io.Reader) *scanner {
	return &scanner{
		r:   bufio.NewReader(r),
		pos: 1,
	}
}

func (s *scanner) read() rune {
	if len(s.buf) > 0 {
		var ch rune
		ch, s.buf = s.buf[0], s.buf[1:]
		s.pos++
		return ch
	}
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	s.pos++
	return ch
}

func (s *scanner) unread(ch rune) {
	s.buf = append(s.buf, ch)
	s.pos--
}

func (s *scanner) scan() (int, token.Token, string) {
	ch := s.read()
	if ch == eof {
		return s.pos, token.EOF, ""
	}
	if s.isReadingIndex {
		switch ch {
		case '\'':
			return s.scanQuoteString('\'')
		case '"':
			return s.scanQuoteString('"')
		case ']':
			s.isReadingIndex = false
			return s.pos - 1, token.RBRACK, "]"
		}
		if isDigit(ch) {
			return s.scanInt(ch)
		}
		return s.pos - 1, token.ILLEGAL, string(ch)
	}
	switch ch {
	case '$':
		return s.pos - 1, token.ROOT, "$"
	case '.':
		return s.pos - 1, token.PERIOD, "."
	case '[':
		s.isReadingIndex = true
		return s.pos - 1, token.LBRACK, "["
	case ']':
		s.isReadingIndex = false
		return s.pos - 1, token.RBRACK, "]"
	}
	return s.scanString(ch)
}

func (s *scanner) scanString(head rune) (int, token.Token, string) {
	var b strings.Builder
	b.WriteRune(head)
scan:
	for {
		ch := s.read()
		switch ch {
		case eof:
			break scan
		case '[', '.':
			s.unread(ch)
			break scan
		}
		b.WriteRune(ch)
	}
	return s.pos - b.Len(), token.STRING, b.String()
}

func (s *scanner) scanQuoteString(term rune) (int, token.Token, string) {
	var b strings.Builder
	var backslashes int
scan:
	for {
		ch := s.read()
		switch ch {
		case eof:
			// string not terminated
			return s.pos, token.ILLEGAL, ""
		case '\\':
			escaped := s.read()
			switch escaped {
			case '\\', term:
				b.WriteRune(escaped)
				backslashes++
			default:
				return s.pos - 2, token.ILLEGAL, ""
			}
		case term:
			break scan
		default:
			b.WriteRune(ch)
		}
	}
	return s.pos - b.Len() - backslashes - 2, token.STRING, b.String()
}

func (s *scanner) scanInt(head rune) (int, token.Token, string) {
	var b strings.Builder
	b.WriteRune(head)
scan:
	for {
		ch := s.read()
		if !isDigit(ch) {
			s.unread(ch)
			break scan
		}
		b.WriteRune(ch)
	}
	if head == '0' && b.Len() != 1 {
		return s.pos - b.Len(), token.ILLEGAL, b.String()
	}
	return s.pos - b.Len(), token.INT, b.String()
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}
