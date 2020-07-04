package parser

import (
	"bufio"
	"bytes"
	xmlreader "github.com/RishabhBhatnagar/gordf/rdfloader/xmlreader"
	"io"
	"testing"
)

func xmlreaderFromString(fileContent string) xmlreader.XMLReader {
	return xmlreader.XMLReaderFromFileObject(bufio.NewReader(io.Reader(bytes.NewReader([]byte(fileContent)))))
}

func TestTriple_Hash(t *testing.T) {
	testTriple := Triple{
		Subject:   &Node{BLANK, ""},
		Predicate: &Node{BLANK, ""},
		Object:    &Node{BLANK, ""},
	}

	expectedHash := "{(BNODE, ); (BNODE, ); (BNODE, )}"
	if expectedHash != testTriple.Hash() {
		t.Errorf("expected %v, found %v", expectedHash, testTriple.Hash())
	}
}

func TestNew(t *testing.T) {
	// testing if the initialized parameters are okay.
	newParser := New()

	// Triples should be initially empty
	if len(newParser.Triples) > 0 {
		t.Errorf("Initialized parser shouldn't have any triple")
	}
}

func TestParser_Parse(t *testing.T) {

	// TestCase 1: only root tag in the document.
	func() {
		emptyValidRDF := `
			<rdf:RDF
				xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
				xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">
			</rdf:RDF>`
		rdfParser := New()
		xmlReader := xmlreaderFromString(emptyValidRDF)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		err = rdfParser.Parse(rootBlock)
		// there shouldn't be any error parsing the content
		if err != nil {
			t.Errorf("unexpected error parsing the document: %v", err)
		}
		// there shouldn't be any triple in the parsed document.
		if len(rdfParser.Triples) != 0 {
			t.Errorf("empty document must have no triples. Found %v", rdfParser.Triples)
		}
	}()

	// TestCase 2:
	// empty rdf with prolog
	func() {
		emptyRDFWithProlog := `<? xml version="1.0" ?>
			<rdf:RDF
			xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
			xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">
			</rdf:RDF>`
		xmlReader := xmlreaderFromString(emptyRDFWithProlog)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		rdfParser := New()
		err = rdfParser.Parse(rootBlock)
		// there shouldn't be any error parsing the content
		if err != nil {
			t.Errorf("unexpected error parsing the document: %v", err)
		}
		// there shouldn't be any triple in the parsed document and
		//     the prolog is not counted in the triples
		if len(rdfParser.Triples) != 0 {
			t.Errorf("empty document must have no triples. Found %v", rdfParser.Triples)
		}
	}()

	// TestCase 3:
	// Invalid RDF with stray characters before closing tag.
	func() {
		invalidRDF := "......<rdf:RDF>"
		xmlReader := xmlreaderFromString(invalidRDF)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		rdfParser := New()
		err = rdfParser.Parse(rootBlock)
		if err == nil {
			t.Errorf("should've raised an error stating stray characters found")
		}
	}()

	// TestCase 4:
	// Valid RDF with single valid triple.
	func() {
		twoTripleRDF := `
			<rdf:RDF
				xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
				xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
				xmlns:example="https://www.sample.com/example">
				<rdf:Description>
					<example:Tag> Name </example:Tag>
				</rdf:Description>
			</rdf:RDF>`
		xmlReader := xmlreaderFromString(twoTripleRDF)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		rdfParser := New()
		err = rdfParser.Parse(rootBlock)
		if err != nil {
			t.Errorf("error parsing a valid rdf file. Error: %v", err)
		}
		if len(rdfParser.Triples) != 2 {
			t.Errorf("expected rdfParser to have exactly two triples. %v triples found", len(rdfParser.Triples))
		}
	}()

	// TestCase 5: extra tag in the rdf.
	func() {
		extraTagRDF := `
			<rdf:RDF
				xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
				xmlns:example="https://www.sample.com/example">
				<rdf:Description>
					<example:Tag> Name </example:Tag>
				</rdf:Description>
				<example:extraTag>
			</rdf:RDF>`
		xmlReader := xmlreaderFromString(extraTagRDF)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		rdfParser := New()
		err = rdfParser.Parse(rootBlock)
		if err == nil {
			t.Errorf("expected an EOF error")
		}
	}()

	// TestCase 6: mismatch tag
	func() {
		invalidRDF := `
			<rdf:RDF
				xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
			</rdf:rdf>`
		xmlReader := xmlreaderFromString(invalidRDF)
		rootBlock, err := xmlReader.Read()
		if err != nil {
			return
		}
		rdfParser := New()
		err = rdfParser.Parse(rootBlock)
		if err == nil {
			t.Errorf("expected an error stating opening and closing tags are not same")
		}
	}()
}
