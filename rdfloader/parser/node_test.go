package parser

import (
	"testing"
)

func TestBlankNodeGetter_Get(t *testing.T) {
	// default blank node getter:
	blankNodeGetter := BlankNodeGetter{}

	// by default, node id starts with N1.
	firstBlankNode := blankNodeGetter.Get()
	if firstBlankNode.ID != "N1" {
		t.Errorf("expected first node to be N1, found %v", firstBlankNode.ID)
	}
	if nodeType := firstBlankNode.NodeType; nodeType != BLANK {
		t.Errorf("blank node must be of type %v. Found %v", BLANK, nodeType)
	}

	secondBlankNode := blankNodeGetter.Get()
	if secondBlankNode.ID != "N2" {
		t.Errorf("expected node to be N2, found %v", secondBlankNode.ID)
	}

	// blank node getter with custom lastid.
	blankNodeGetter = BlankNodeGetter{
		lastid: -1,
	}
	// last id -1 means that first node should start from N0
	firstBlankNode = blankNodeGetter.Get()
	if firstBlankNode.ID != "N0" {
		t.Errorf("Expected node to be %v, found %v", "N0", firstBlankNode.ID)
	}
}

func TestBlankNodeGetter_GetFromId(t *testing.T) {
	// default blank node getter:
	getter := BlankNodeGetter{} // id starts with 1
	blankNode := getter.Get()

	if blankNode.ID != "N1" {
		t.Errorf("expected first node's id to be N1, found %v", blankNode.ID)
	}

	blankNodeA0 := getter.GetFromId("A0")
	if blankNodeA0.ID != "NA0" {
		t.Errorf("expected first node's id to be NA0, found %v", blankNodeA0.ID)
	}
}

func TestNode_String(t *testing.T) {
	// a node is wrapped in a tuple with (NodeType, ID)
	node := Node{
		BLANK,
		"",
	}
	if node.String() != "(BNODE, )" {
		t.Errorf("String representation of blank node with empty id should be (BNODE, ). Found %v", node.String())
	}
}
