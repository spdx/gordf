package parser

import "fmt"

type NODETYPE string

const (
	LITERAL NODETYPE = "LITERAL"
	BLANK            = "BNODE"
	IRI              = "IRI"
)

type Node struct {
	NodeType NODETYPE
	ID      string
}
func (node *Node) String() string {
	return fmt.Sprintf("(%v, %v)", node.NodeType, node.ID)
}

type BlankNodeGetter struct {
	lastid int
}

func (getter *BlankNodeGetter) Get() Node {
	getter.lastid += 1
	return Node {
		NodeType: BLANK,
		ID:      fmt.Sprintf("N%v", getter.lastid),
	}
}

func (getter *BlankNodeGetter) GetFromId(id string) Node {
	return Node {
		NodeType: BLANK,
		ID: fmt.Sprintf("N%v", id),
	}
}