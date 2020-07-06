package parser

import (
	"fmt"
	xmlreader "github.com/RishabhBhatnagar/gordf/rdfloader/xmlreader"
	"github.com/RishabhBhatnagar/gordf/uri"
)

type Triple struct {
	Subject, Predicate, Object *Node
}

func (parser *Parser) appendTriple(triple *Triple) {
	// does what it say.
	// appends the triples to the parser.
	// uses a lock for mutex to prevent race condition.

	// writelock is a type of RWMutex.
	// RW is Readers-Writer Lock. That is, at a time, more than on e readers
	//		can read from the data structure but at a time only one writer can
	//		access the data structure.
	parser.writeLock.Lock()
	if _, exists := parser.setTriples[triple.Hash()]; !exists {
		// append to the map and triples' set if it doesn't already exist in the map.
		parser.setTriples[triple.Hash()] = triple
		parser.Triples = append(parser.Triples, triple)
	}
	parser.writeLock.Unlock()
}

func (parser *Parser) uriFromPair(schemaName, name string) (mergedUri uri.URIRef, err error) {
	// returns the uri representation of a pair of strings.
	// name:schemaName is an example of pair.
	// pairs such as rdf:RDF, where, rdf must be a valid xmlns schema name.

	// base must be a valid schema name defined in the root tag.
	baseURI, ok := parser.schemaDefinition[schemaName]
	if !ok {
		return uri.URIRef{}, fmt.Errorf("undefined schema name: %v", schemaName)
	}

	// adding the relative fragment to the base uri.
	return baseURI.AddFragment(name), nil
}

func (parser *Parser) nodeFromTag(openingTag xmlreader.Tag) (node *Node, err error) {
	// returns the node object from the opening tag of any block.
	// https://www.w3.org/TR/rdf-syntax-grammar/figure1.png has sample image having 5 nodes.
	// 		one of them is a blank node.

	// description of the entire function:
	// if the opening tag has an attribute of rdf:about,
	//		the node will represented by the value of rdf:about attribute
	// else, it is a blank node.

	// checking if any of the attributes is a rdf:about attribute
	index, err := parser.getRDFAttributeIndex(openingTag, "about")
	if err != nil {
		return
	}

	var currentNode Node
	if index == -1 {
		// we didnt' find rdf:about in the attributes of the opening tag.
		// returning a new blank node.
		rdfNodeIDIndex, err := parser.getRDFAttributeIndex(openingTag, "nodeID")
		if err != nil {
			return nil, err
		}
		if rdfNodeIDIndex == -1 {
			currentNode = parser.blankNodeGetter.Get()
		} else {
			currentNode = parser.blankNodeGetter.GetFromId(openingTag.Attrs[rdfNodeIDIndex].Value)
		}
	} else {
		// we found a rdf:about tag.
		currentNode = Node{
			NodeType: IRI,
			ID:       openingTag.Attrs[index].Value,
		}
	}
	return &currentNode, nil
}

func (triple *Triple) Hash() string {
	return fmt.Sprintf("{%v; %v; %v}", triple.Subject, triple.Predicate, triple.Object)
}
