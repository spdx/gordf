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
	subject, predicate, object Node
}

func New() (parser *Parser) {
	// creates a new parser object
	return &Parser{
		Triples:          []*Triple{},
		schemaDefinition: map[string]uri.URIRef{},
		blankNodeGetter:  BlankNodeGetter{-1},
	}
}

func (parser *Parser) appendTriple(t *Triple) {
	parser.Triples = append(parser.Triples, t)
}

func (parser *Parser) uriFromPair(schemaName string, name string) (uri.URIRef, error) {
	if len(schemaName) == 0 {
		// no schema name given.
		return uri.NewURIRef(name)
	}
	schemaURI, exists := parser.schemaDefinition[schemaName]
	if !exists {
		// schema name doesn't exist in the root block.
		return uri.URIRef{}, fmt.Errorf("undefined schema name: %v", schemaName)
	}
	return schemaURI.AddFragment(name), nil
}

func (parser *Parser) parseAttributes(parentNode Node, attributes []xmlreader.Attribute) error {
	for _, attr := range attributes {
		if !(attr.Name == "about" || attr.Name == "id") {
			predicateURI, err := parser.uriFromPair(attr.SchemaName, attr.Name)
			if err != nil {
				return err
			}
			parser.appendTriple(&Triple{
				parentNode,
				Node{IRI, predicateURI.String()},
				Node{IRI, attr.Value},
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
			parentNode,
			openingTagNode,
			Node{LITERAL, block.Value},
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

func (parser *Parser) Parse(filePath string) error {
	reader, err := xmlreader.XMLReaderFromFilePath(filePath)
	if err != nil {
		return err
	}

	rootBlock, err := reader.Read()
	if err != nil {
		return err
	}

	schemaDefinition, err := parseHeaderBlock(rootBlock)
	if err != nil {
		return err
	}
	parser.schemaDefinition = schemaDefinition

	rootNode := parser.blankNodeGetter.Get()
	for _, childBlock := range rootBlock.Children {
		err = parser.parseChild(rootNode, childBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseHeaderBlock(rootBlock xmlreader.Block) (map[string]uri.URIRef, error) {
	namespaceURI := map[string]uri.URIRef{}
	for _, attr := range rootBlock.OpeningTag.Attrs {
		if attr.SchemaName == "xmlns" {
			uriref, err := uri.NewURIRef(attr.Value)
			if err != nil {
				err = fmt.Errorf("schema URI %v doesn't confirm to URL rules", rootBlock)
				return namespaceURI, err
			}
			namespaceURI[attr.Name] = uriref
		}
	}
	return namespaceURI, nil
}
