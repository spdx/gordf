package parser

import (
	"fmt"
	xmlreader "github.com/RishabhBhatnagar/gordf/rdfloader/xmlreader"
	"github.com/RishabhBhatnagar/gordf/uri"
	"sync"
)

const RDFNS = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"

type Parser struct {
	Triples          []*Triple
	schemaDefinition map[string]uri.URIRef
	blankNodeGetter  BlankNodeGetter
	rdfNS uri.URIRef
	wg sync.WaitGroup
}

type Triple struct {
	Subject, Predicate, Object *Node
}


func parseHeaderBlock(rootBlock xmlreader.Block) (map[string]uri.URIRef, error) {
	// returns all the schema definitions in the root block.
	// a schema definition is of the form xmlns:SchemaName="URI",

	namespaceURI := map[string]uri.URIRef{}

	for _, attr := range rootBlock.OpeningTag.Attrs {
		if attr.SchemaName == "xmlns" {
			uriref, err := uri.NewURIRef(attr.Value)
			if err != nil {
				return namespaceURI, fmt.Errorf("schema URI %v doesn't confirm to URL rules", rootBlock)
			}
			namespaceURI[attr.Name] = uriref
		}
	}
	return namespaceURI, nil
}


func (parser *Parser) uriFromPair(name, schemaName string) (mergedUri uri.URIRef, err error) {
	// returns the uri representation of a pair of strings.
	// name:schemaName is an example of pair.
	// pairs such as rdf:RDF, where, rdf must be a valid xmlns schema name.

	// base must be a valid schema name defined in the root tag.
	baseURI, ok := parser.schemaDefinition[name]
	if !ok {
		return uri.URIRef{}, fmt.Errorf("undefined schema name: %v", name)
	}

	// adding the relative fragment to the base uri.
	return baseURI.AddFragment(schemaName), nil
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
	index := -1
	for i, attr := range openingTag.Attrs {
		attrUri, err := parser.uriFromPair(attr.SchemaName, attr.Name)
		if err != nil { return }
		if attrUri == parser.rdfNS.AddFragment("about") {
			// current attribute is a rdf:about tag,
			index = i
			break
		}
	}

	if index == -1 {
		// we didnt' find rdf:about in the attributes of the opening tag.
		// returnning a new blank node.
		blankNode := parser.blankNodeGetter.Get()
		return &blankNode, nil
	}

	// we found a rdf:about tag.
	currentNode := Node{
		NodeType: IRI,
		Val: openingTag.Attrs[index].Value,
	}
	return &currentNode, nil
}


func New() (parser *Parser) {
	// creates a new parser object
	rdfNS, _ := uri.NewURIRef(RDFNS)
	return &Parser{
		Triples:          []*Triple{},
		schemaDefinition: map[string]uri.URIRef{"": uri.URIRef{}},
		blankNodeGetter:  BlankNodeGetter{-1},
		wg: sync.WaitGroup{},
		rdfNS: rdfNS,
	}
}


func (parser *Parser) parseBlock(block *xmlreader.Block) (err error) {
	/*
	1. What is a block?
	Ans: A rdf block is made up of
			1. Root Node (IRI Ref or BlankNode) :: Subject
			2. Link (IRI Ref)                   :: Object
			3. anotherBlock (Literal or IRI Ref or Blank Node) :: Predicate

	2. Example of a Block.
		Sample RDF/XML input with non-blank subject and literal predicate.
			<spdx:License rdf:about="http://spdx.org/licenses/Apache-2.0">
				<spdx:licenseId>Apache-2.0</spdx:licenseId>
			</spdx:License>
		Output Components:
			Subject:   http://spdx.org/licenses/Apache-2.0  (IRI Ref)
			Object:    spdx:licenseId						(IRI Ref)
			Predicate: Apacha-2.0							(Literal)

		If the rdf:about attribute of the subject is removed, it will become a blank node.
	*/

}

func (parser *Parser) Parse(filePath string) error {
	// reader for xml file
	reader, err := xmlreader.XMLReaderFromFilePath(filePath)
	if err != nil {
		return err
	}
	// parsing the xml file
	rootBlock, err := reader.Read()
	if err != nil {
		return err
	}

	// set all the schema definitions in the root block.
	schemaDefinition, err := parseHeaderBlock(rootBlock)
	if err != nil {
		return err
	}
	parser.schemaDefinition = schemaDefinition

	// root tag is set now.
	for _, child := range rootBlock.Children {
		err = parser.parseBlock(nil, nil, child)
		if err != nil { return err }
	}

	return nil
}
