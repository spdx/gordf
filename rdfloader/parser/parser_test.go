package parser

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

func newTestFile(content string) (filenames string, destructor func(), err error) {
	rand.Seed(time.Now().UnixNano()) // to ensure a pseudo random number every time.
	fileName := "!#$^$" + string(rand.Int())
	err = ioutil.WriteFile(fileName, []byte(content), 777)
	if err != nil {
		return
	}

	return fileName, func() {
		if _, err = os.Stat(fileName); err == nil {
			// file exists
			os.Remove(fileName)
		}
	}, nil
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
	emptyValidRDF := `
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">
	</rdf:RDF>
	`
	filename, destructor, err := newTestFile(emptyValidRDF)
	if err != nil {
		t.Errorf(err.Error())
	}
	rdfParser := New()
	err = rdfParser.Parse(filename)
	// there shouldn't be any error parsing the content
	if err != nil {
		t.Errorf("unexpected error parsing the document: %v", err)
	}
	// there shouldn't be any triple in the parsed document.
	if len(rdfParser.Triples) != 0 {
		t.Errorf("empty document must have no triples. Found %v", rdfParser.Triples)
	}
	destructor() // deletes the temporary file

	// TestCase 2:
	prolog := `<? xml version="1.0" ?>`
	emptyRDFWithProlog := prolog + emptyValidRDF
	filename, destructor, err = newTestFile(emptyRDFWithProlog)
	if err != nil {
		t.Errorf(err.Error())
	}
	rdfParser = New() //  reinitialize the parser.
	err = rdfParser.Parse(filename)
	// there shouldn't be any error parsing the content
	if err != nil {
		t.Errorf("unexpected error parsing the document: %v", err)
	}
	// there shouldn't be any triple in the parsed document and
	//     the prolog is not counted in the triples
	if len(rdfParser.Triples) != 0 {
		t.Errorf("empty document must have no triples. Found %v", rdfParser.Triples)
	}
	destructor()

	// TestCase 3:
	// Invalid RDF with stray characters before closing tag.
	invalidRDF := "......" + emptyValidRDF
	filename, destructor, err = newTestFile(invalidRDF)
	if err != nil {
		t.Errorf("error creating a test file: %v", err)
	}
	rdfParser = New()
	err = rdfParser.Parse(filename)
	if err == nil {
		t.Errorf("should've raised an error stating stray characters found")
	}
	destructor()

	// TestCase 4:
	// Valid RDF with single valid triple.
	twoTripleRDF := `
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
		xmlns:example="https://www.sample.com/example">
		<rdf:Description>
			<example:Tag> Name </example:Tag>
		</rdf:Description>
	</rdf:RDF>
	`
	filename, destructor, err = newTestFile(twoTripleRDF)
	if err != nil {
		t.Errorf("error creating a test file: %v", err)
	}
	rdfParser = New()
	err = rdfParser.Parse(filename)
	if err != nil {
		t.Errorf("error parsing a valid rdf file. Error: %v", err)
	}
	if len(rdfParser.Triples) != 2 {
		t.Errorf("expected rdfParser to have exactly two triples. %v triples found", len(rdfParser.Triples))
	}
	destructor()

	// TestCase 5: extra tag in the rdf.
	extraTagRDF := `
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns:example="https://www.sample.com/example">
		<rdf:Description>
			<example:Tag> Name </example:Tag>
		</rdf:Description>
		<example:extraTag>
	</rdf:RDF>`
	filename, destructor, _ = newTestFile(extraTagRDF)
	rdfParser = New()
	err = rdfParser.Parse(filename)
	if err == nil {
		t.Errorf("expected an EOF error")
	}
	destructor()

	// TestCase 6: mismatch tag
	invalidRDF = `<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
	</rdf:rdf>`
	filename, destructor, _ = newTestFile(invalidRDF)
	rdfParser = New()
	err = rdfParser.Parse(filename)
	if err == nil {
		t.Errorf("expected an error stating opening and closing tags are not same")
	}
	destructor()
}
