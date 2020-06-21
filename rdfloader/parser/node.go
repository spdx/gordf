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
	Val      string
}


type BlankNodeGetter struct {
	lastid int
}


func (getter *BlankNodeGetter) Get() Node {
	getter.lastid += 1
	return Node{
		NodeType: BLANK,
		Val:      fmt.Sprintf("N%v", getter.lastid),
	}
}