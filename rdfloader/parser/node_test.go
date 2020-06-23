package parser

import (
	"testing"
)

func TestBlankNodeGetter_Get(t *testing.T) {
	// default blank node getter:
	blankNodeGetter := BlankNodeGetter{}

	// by default, node id starts with N1.
	firstBlankNode := blankNodeGetter.Get()
	if firstBlankNode.Val != "N1" {
		t.Errorf("expected first node to be N1, found %v", firstBlankNode.Val)
	}
	if nodeType := firstBlankNode.NodeType; nodeType != BLANK {
		t.Errorf("blank node must be of type %v. Found %v", BLANK, nodeType)
	}

	secondBlankNode := blankNodeGetter.Get()
	if secondBlankNode.Val != "N2" {
		t.Errorf("expected node to be N2, found %v", secondBlankNode.Val)
	}

	// blank node getter with custom lastid.
	blankNodeGetter = BlankNodeGetter{
		lastid: -1,
	}
	// last id -1 means that first node should start from N0
	firstBlankNode = blankNodeGetter.Get()
	if firstBlankNode.Val != "N0" {
		t.Errorf("Expected node to be %v, found %v", "N0", firstBlankNode.Val)
	}
}
