package parser

import (
	"fmt"
	xmlreader "github.com/RishabhBhatnagar/gordf/rdfloader/xmlreader"
	"github.com/RishabhBhatnagar/gordf/uri"
)

const RDFNS = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"

type Parser struct {
	Triples          []*Triple
	schemaDefinition map[string]uri.URIRef
	blankNodeGetter  BlankNodeGetter
}

type Triple struct {
	Subject, Predicate, Object Node
}


func (parser *Parser) appendTriple(t *Triple) {
	parser.Triples = append(parser.Triples, t)
}

// returns the schema-uri corresponding to the <schemaName:name> tag.
func (parser *Parser) uriFromPair(schemaName string, name string) (uri.URIRef, error) {
	if len(schemaName) == 0 {
		// no schema name given.
		return uri.NewURIRef(name)
	}

	// checking if the schemaName is declared in the root.
	schemaURI, exists := parser.schemaDefinition[schemaName]
	if !exists {
		// schema name doesn't exist in the root block.
		return uri.URIRef{}, fmt.Errorf("undefined schema name: %v", schemaName)
	}
	return schemaURI.AddFragment(name), nil
}

func (parser *Parser) parseAttributes(parentNode Node, attributes []xmlreader.Attribute) error {

	// searching for rdf:about or rdf:id in all the attributes
	tagUriString := ""
	for _, attr := range attributes {
		predicateURI, err := parser.uriFromPair(attr.SchemaName, attr.Name)
		if err != nil { return err }
		if predicateURI.String() == RDFNS + "about" || predicateURI.String() == RDFNS + "id" {
			tagUriString = attr.Value
			break
		}
	}

	// if rdf:about was found, subject of the current tag must be the value of the rdf:about tag.
	if len(tagUriString) > 0 {
		tagUri, err := uri.NewURIRef(tagUriString)
		if err != nil { return err }
		parentNode = Node {IRI, tagUri.String()}
	}

	for _, attr := range attributes {
		predicateURI, err := parser.uriFromPair(attr.SchemaName, attr.Name)
		if err != nil { return err }
		if !(predicateURI.String() == RDFNS + "about" || predicateURI.String() == RDFNS + "id") {
			parser.appendTriple(&Triple{
				Subject:   parentNode,
				Predicate: Node{IRI, predicateURI.String()},
				Object:    Node{IRI, attr.Value},
			})
		}
	}
	return nil
}


func (parser *Parser) parseChild(parentNode Node, block *xmlreader.Block) (err error) {
	name, schemaName := block.OpeningTag.Name, block.OpeningTag.SchemaName
	openingTagURI, err := parser.uriFromPair(schemaName, name)
	if err != nil {
		return err
	}
	openingTagNode := Node{IRI, openingTagURI.String()}
	if len(block.Children) == 0 {
		// value attribute must be set.
		parser.appendTriple(&Triple{
			Subject:   parentNode,
			Predicate: openingTagNode,
			Object:    Node{LITERAL, block.Value},
		})
	} else {
		// there are children of current node.
		for _, childBlock := range block.Children {
			err = parser.parseChild(openingTagNode, childBlock)
			if err != nil {
				return
			}
		}
	}
	return parser.parseAttributes(openingTagNode, block.OpeningTag.Attrs)

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


func New() (parser *Parser) {
	// creates a new parser object
	return &Parser{
		Triples:          []*Triple{},
		schemaDefinition: map[string]uri.URIRef{},
		blankNodeGetter:  BlankNodeGetter{-1},
	}
}


func (parser *Parser) Parse(filePath string) error {
	// reader for xml file
	reader, err := xmlreader.XMLReaderFromFilePath(filePath)
	if err != nil {
		return err
	}
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

	// root node is always identified by a blank node.
	rootNode := parser.blankNodeGetter.Get()

	// parse each child of the root block.
	for _, childBlock := range rootBlock.Children {
		err = parser.parseChild(rootNode, childBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
