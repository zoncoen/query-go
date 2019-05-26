// Package token defines constants representing the lexical tokens.
package token

// Token is the set of lexical tokens.
type Token int

const (
	ILLEGAL Token = iota
	EOF

	STRING // "text"
	INT    // 123

	PERIOD // .
	LBRACK // [
	RBRACK // ]
)

// String returns t as string.
func (t Token) String() string {
	switch t {
	case EOF:
		return "EOF"
	case STRING:
		return "string"
	case INT:
		return "int"
	case PERIOD:
		return "period"
	case LBRACK:
		return "lbrack"
	case RBRACK:
		return "rbrack"
	}
	return "illegal"
}
