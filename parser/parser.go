// Package parser implements a parser for a query string.
package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/zoncoen/query-go/ast"
	"github.com/zoncoen/query-go/token"
)

// Parser represents a parser.
type Parser struct {
	s      *scanner
	pos    int
	tok    token.Token
	lit    string
	errors Errors
}

// NewParser returns a new parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: newScanner(r)}
}

// Parse parses the query string and returns the corresponding ast.Node.
func (p *Parser) Parse() (ast.Node, error) {
	return p.parse(), p.errors.Err()
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.s.scan()
}

func (p *Parser) parse() ast.Node {
	node := p.parseFirst()
	if node == nil {
		return nil
	}
L:
	for {
		switch p.tok {
		case token.PERIOD:
			pos := p.pos
			p.next()
			node = &ast.Selector{
				ValuePos: pos,
				X:        node,
				Sel:      p.lit,
			}
			p.next()
		case token.LBRACK:
			node = p.parseIndex(node)
		case token.EOF:
			break L
		default:
			p.expect(token.PERIOD, token.LBRACK)
		}
	}
	return node
}

func (p *Parser) parseFirst() ast.Node {
	p.next()
	var node ast.Node
	switch p.tok {
	case token.ROOT:
		node = &ast.Root{
			ValuePos: p.pos,
		}
		p.next()
	case token.STRING:
		node = &ast.Selector{
			ValuePos: p.pos,
			Sel:      p.lit,
		}
		p.next()
	case token.PERIOD:
		pos := p.pos
		p.next()
		node = &ast.Selector{
			ValuePos: pos,
			X:        node,
			Sel:      p.lit,
		}
		p.next()
	case token.LBRACK:
		node = p.parseIndex(nil)
	case token.EOF:
		return nil
	default:
		p.expect(token.ROOT, token.STRING, token.LBRACK)
	}
	return node
}

func (p *Parser) parseIndex(x ast.Node) ast.Node {
	pos := p.pos
	p.next()
	var node ast.Node
	switch p.tok {
	case token.STRING:
		node = &ast.Selector{
			ValuePos: pos,
			X:        x,
			Sel:      p.lit,
		}
		p.next()
	case token.INT:
		node = &ast.Index{
			ValuePos: pos,
			X:        x,
			Index:    p.parseInt(),
		}
	default:
		p.expect(token.STRING, token.INT)
	}
	p.expect(token.RBRACK)
	return node
}

func (p *Parser) parseInt() int {
	pos, lit := p.pos, p.lit
	p.next()
	i, err := strconv.Atoi(lit)
	if err != nil {
		p.error(pos, fmt.Sprintf("%s is not an integer", lit))
		return 0
	}
	return i
}

func (p *Parser) error(pos int, msg string) {
	p.errors.Append(pos, msg)
}

func (p *Parser) expect(toks ...token.Token) {
	var ok bool
	strs := make([]string, len(toks))
	for i, tok := range toks {
		if p.tok == tok {
			ok = true
			break
		}
		strs[i] = fmt.Sprintf(`"%s"`, tok)
	}
	if !ok {
		p.error(p.pos, fmt.Sprintf(`expected %s but found "%s"`, strings.Join(strs, " or "), p.tok))
	}
	p.next() // make progress
}
