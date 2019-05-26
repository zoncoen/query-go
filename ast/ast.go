// Package ast declares the types used to represent syntax trees.
package ast

// All node types implement the Node interface.
type Node interface {
	Pos() int
}

type (
	// A Selector node represents an expression followed by a selector.
	Selector struct {
		ValuePos int
		X        Node
		Sel      string
	}

	// An Index node represents an expression followed by an index.
	Index struct {
		ValuePos int
		X        Node
		Index    int
	}
)

// Pos returns the position of first character belonging to the node.
func (e *Selector) Pos() int { return e.ValuePos }
func (e *Index) Pos() int    { return e.ValuePos }
