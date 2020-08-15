package rdfwriter

import (
	"bytes"
	"github.com/RishabhBhatnagar/gordf/rdfloader/parser"
	"github.com/RishabhBhatnagar/gordf/uri"
	"reflect"
	"testing"
)

// returns a schemaDefinition with spdx and rdf uris.
func getSampleSchemaDefinition() map[string]uri.URIRef {
	schemaDefinition := make(map[string]uri.URIRef)

	// adding rdf uri:
	uriref, _ := uri.NewURIRef("http://www.w3.org/1999/02/22-rdf-syntax-ns#")
	schemaDefinition["rdf"] = uriref

	// adding the spdx uri:
	uriref, _ = uri.NewURIRef("http://spdx.org/rdf/terms#")
	schemaDefinition["spdx"] = uriref

	return schemaDefinition
}

func TestFilterTriples(t *testing.T) {
	nodes := getNBlankNodes(10)
	triples := []*parser.Triple{
		{nodes[0], nodes[1], nodes[2]},
		{nodes[3], nodes[1], nodes[4]},
		{nodes[3], nodes[5], nodes[4]},
	}

	// TestCase 1: Default Filtering: must return all the triples
	triplesAfterFiltering := FilterTriples(triples, nil, nil, nil)
	if len(triples) != len(triplesAfterFiltering) {
		t.Errorf("default filtering with all filter params nil didn't return all the triples as the output")
	}

	// TestCase 2: Filtering on subject:
	triplesAfterFiltering = FilterTriples(triples, &nodes[3].ID, nil, nil)
	if !reflect.DeepEqual(triplesAfterFiltering, triples[1:]) {
		t.Errorf("subject filtering faulty")
	}

	// TestCase 3: Filtering on predicate
	triplesAfterFiltering = FilterTriples(triples, nil, &nodes[1].ID, nil)
	if !reflect.DeepEqual(triplesAfterFiltering, triples[:2]) {
		t.Errorf("predicate filtering faulty")
	}

	// TestCase 4: Filtering on object
	triplesAfterFiltering = FilterTriples(triples, nil, nil, &nodes[4].ID)
	if !reflect.DeepEqual(triplesAfterFiltering, triples[1:]) {
		t.Errorf("object filtering faulty")
	}

	// TestCase 4: Filtering on subject and object both
	triplesAfterFiltering = FilterTriples(triples, &nodes[3].ID, nil, &nodes[4].ID)
	if !reflect.DeepEqual(triplesAfterFiltering, triples[1:]) {
		t.Errorf("faulty filtering for 2 attributes at a time")
	}

	// TestCase 5: Filtering on predicate
	triplesAfterFiltering = FilterTriples(triples, nil, &nodes[1].ID, nil)
	if !reflect.DeepEqual(triplesAfterFiltering, triples[:2]) {
		t.Errorf("predicate filtering faulty")
	}
}

func TestTriplesToString(t *testing.T) {
	// init all required variables for the testing.
	var triples []*parser.Triple
	schemaDefinition := getSampleSchemaDefinition()
	tab := "    " // 4 spaces as a tab character
	bnodes := getNBlankNodes(5)

	// TestCase 1: error raised by stringify must be returned by the function too.
	// stringify will complain that every subject node must be associated with
	// at least one triple with predicate rdf:type
	triples = append(triples, &parser.Triple{
		Subject:   bnodes[0],
		Predicate: bnodes[1],
		Object:    bnodes[2],
	})
	_, err := TriplesToString(triples, schemaDefinition, tab)
	if err == nil {
		t.Errorf("expected an error stating invalid triples")
	}

	// TestCase 2: valid input without any triples must return just the root tag with the schemaDefinition.
	// clearing the schemaDefinition such that the function returns just the root tags.
	triples = nil // clearing the slice.
	schemaDefinition = nil
	output, err := TriplesToString(triples, schemaDefinition, tab)
	expectedOutput := `<rdf:RDF>
</rdf:RDF>`
	if output != expectedOutput {
		t.Errorf("expected output: %s, got output: %s", expectedOutput, output)
	}

	// TestCase 3: simple testcase with more than one triples.
	schemaDefinition = getSampleSchemaDefinition()
	spdxRef := schemaDefinition["spdx"]
	triples = []*parser.Triple{
		{
			Subject:   bnodes[0],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
			Object:    &parser.Node{NodeType: parser.IRI, ID: spdxRef.String() + "Snippet"},
		},
		{
			Subject:   bnodes[0],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: spdxRef.String() + "randomFragment"},
			Object:    &parser.Node{NodeType: parser.LITERAL, ID: "sample"},
		},
	}
	output, err = TriplesToString(triples, schemaDefinition, tab)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWriteToFile(t *testing.T) {
	// init all required variables for the testing.
	var triples []*parser.Triple
	schemaDefinition := getSampleSchemaDefinition()
	tab := "    " // 4 spaces as a tab character
	bnodes := getNBlankNodes(10)

	// TestCase 1: checking if the function returns an error when the
	//             TriplesToString function raises an error.
	// expected error: every subject node must be associated with exactly 1
	//                 triple of type rdf:type predicate.
	triples = append(triples, &parser.Triple{
		Subject:   bnodes[0],
		Predicate: bnodes[1],
		Object:    bnodes[2],
	})
	var b bytes.Buffer
	err := WriteToFile(&b, triples, schemaDefinition, tab)
	if err == nil {
		t.Errorf("expected an error stating invalid triples")
	}

	// TestCase 2: valid case with no triples and no schema definitions in the root tag.
	triples = nil
	schemaDefinition = nil
	b = bytes.Buffer{}
	expectedOutput := `<rdf:RDF>
</rdf:RDF>`
	err = WriteToFile(&b, triples, schemaDefinition, tab)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if b.String() != expectedOutput {
		t.Errorf("wrong output written to the buffer. expected %s, found %s", expectedOutput, b.String())
	}
}

func Test_stringify(t *testing.T) {
	bnodes := getNBlankNodes(10)
	var triples []*parser.Triple
	nodeToTriples := GetNodeToTriples(triples)
	schemaDefinition := getSampleSchemaDefinition()
	invSchemaDefinition := invertSchemaDefinition(schemaDefinition)
	depth := 0
	tab := "  " // 2 spaces as the tabs

	// TestCase 1: Invalid case with no triples belonging to the passed node.
	// expected error:  every subject node must be associated with exactly 1
	//                  triple of type rdf:type predicate
	_, err := stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	if err == nil {
		t.Errorf("expected error stating nodes must have a triple with predicate of rdf:type")
	}

	// TestCase 2: invalid base uri in the object must return an error
	triples = append(triples, &parser.Triple{
		Subject:   bnodes[0],
		Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
		Object:    &parser.Node{NodeType: parser.IRI, ID: "https://inexistent.com/uri#fragment"},
	})
	nodeToTriples = GetNodeToTriples(triples)
	_, err = stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	if err == nil {
		t.Errorf("expeected an error saying uri not defined in the schemaDefinition")
	}

	// TestCase 3: valid input with only rdf:type triple
	triples[0].Object.ID = "http://spdx.org/rdf/terms#Snippet"
	output, _ := stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	expectedOutput := `<spdx:Snippet>

</spdx:Snippet>`
	if output != expectedOutput {
		t.Errorf("output is not correct. expected output is %s. found %s", expectedOutput, output)
	}

	// TestCase 4: invalid input with two triples: rdf:type and second invalid triples
	triples = append(triples, &parser.Triple{
		Subject:   bnodes[0],
		Predicate: &parser.Node{NodeType: parser.IRI, ID: ""},
		Object:    &parser.Node{NodeType: parser.LITERAL, ID: "comment"},
	})
	nodeToTriples = GetNodeToTriples(triples)
	_, err = stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	if err == nil {
		t.Errorf("expected an error saying invalid predicate uri")
	}

	// TestCase 4: input with two triples: rdf:type and rdf:resource
	triples[1].Predicate.ID = "http://spdx.org/rdf/terms#algorithm"
	triples[1].Object = &parser.Node{
		NodeType: parser.RESOURCELITERAL,
		ID:       "http://spdx.org/rdf/terms#checksumAlgorithm_sha256",
	}
	nodeToTriples = GetNodeToTriples(triples)
	output, _ = stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	expectedOutput = `<spdx:Snippet>
  <spdx:algorithm rdf:resource="http://spdx.org/rdf/terms#checksumAlgorithm_sha256"/>
</spdx:Snippet>`
	if output != expectedOutput {
		t.Errorf("mismatching outputs. Expected:\n%v\n Found: \n%v", expectedOutput, output)
	}

	// TestCase 5: input with two triples: rdf:type and rdf:comment and second triple of object-type literal
	triples[1] = &parser.Triple{
		Subject: bnodes[0],
		Predicate: &parser.Node{
			NodeType: parser.IRI,
			ID:       parser.RDFNS + "comment",
		},
		Object: &parser.Node{
			NodeType: parser.LITERAL,
			ID:       "comment",
		},
	}
	nodeToTriples = GetNodeToTriples(triples)
	_, err = stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// TestCase 6: input with 3 level deep nested triples.
	spdxRef := schemaDefinition["spdx"]
	triples = []*parser.Triple{
		{ // N1 rdf:type spdx:externalRef
			Subject: bnodes[0],
			Predicate: &parser.Node{
				NodeType: parser.IRI,
				ID:       parser.RDFNS + "type",
			},
			Object: &parser.Node{
				NodeType: parser.IRI,
				ID:       spdxRef.String() + "externalRef",
			},
		},
		{ // N1 spdx:ExternalRef N2
			Subject: bnodes[0],
			Predicate: &parser.Node{
				NodeType: parser.IRI,
				ID:       spdxRef.String() + "ExternalRef",
			},
			Object: bnodes[1],
		},
		{ // N2 rdf:type spdx:referenceType
			Subject: bnodes[1],
			Predicate: &parser.Node{
				NodeType: parser.IRI,
				ID:       parser.RDFNS + "type",
			},
			Object: &parser.Node{
				NodeType: parser.IRI,
				ID:       spdxRef.String() + "referenceType",
			},
		},
		{ // N2 spdx:ReferenceType "http://spdx.org/rdf/references/cpe23Type"
			Subject: bnodes[1],
			Predicate: &parser.Node{
				NodeType: parser.IRI,
				ID:       spdxRef.String() + "ReferenceType",
			},
			Object: &parser.Node{
				NodeType: parser.LITERAL,
				ID:       "http://spdx.org/rdf/references/cpe23Type",
			},
		},
	}
	nodeToTriples = GetNodeToTriples(triples)
	output, _ = stringify(bnodes[0], nodeToTriples, invSchemaDefinition, depth, tab)
	expectedOutput = `<spdx:externalRef>
  <spdx:ExternalRef>
    <spdx:referenceType>
      <spdx:ReferenceType>
        http://spdx.org/rdf/references/cpe23Type
      </spdx:ReferenceType>
    </spdx:referenceType>
  </spdx:ExternalRef>
</spdx:externalRef>`
	if output != expectedOutput {
		t.Errorf("mismatching outputs. Expected:\n%v\n Found: \n%v", expectedOutput, output)
	}
}

func Test_getOpeningAndClosingTags(t *testing.T) {
	nodes := getNBlankNodes(10)
	tab := ""
	rdfNSAbbrev := "rdf"
	invSchemaDefinition := map[string]string{
		"http://www.w3.org/1999/02/22-rdf-syntax-ns": "rdf",
		"http://spdx.org/rdf/terms":                  "spdx",
	}
	var triples []*parser.Triple

	// TestCase 1: empty triple list must return an error
	_, _, err := getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	if err == nil {
		t.Errorf("function should've raised an error stating every node must be associated with a triple of predicate rdf:type")
	}

	// TestCase 2: exactly one triple of predicate rdf:type but the object uri is invalid.
	//             Must raise an error
	triples = append(triples, &parser.Triple{
		Subject:   nodes[0],
		Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
		Object:    nodes[2],
	})
	_, _, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	if err == nil {
		t.Errorf("expected an error saying invalid object uri")
	}

	// TestCase 3: exactly one triple of predicate rdf:type with valid object uri
	triples[0].Object = &parser.Node{NodeType: parser.IRI, ID: "http://spdx.org/rdf/terms#Snippet"}
	openingTag, closingTag, err := getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	expectedOpeningTag, expectedClosingTag := "<spdx:Snippet>", "</spdx:Snippet>"
	if openingTag != expectedOpeningTag {
		t.Errorf("wrong opening tag. expected %s, got %s", expectedOpeningTag, openingTag)
	}
	if closingTag != expectedClosingTag {
		t.Errorf("wrong closing tag. expected %s, got %s", expectedClosingTag, closingTag)
	}

	// TestCase 3: more than one triple of type rdf:type
	triples = append(triples, triples[0])
	_, _, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	if err == nil {
		t.Error("function should raise an error when there are more than one triples having rdf:type predicate")
	}

	// resetting triples with only one rdf:type attribute
	triples = triples[:1]

	// TestCase 4: exactly one nodeID attribute (a valid case)
	triples = append(triples, &parser.Triple{
		Subject:   nodes[0],
		Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "nodeID"},
		Object:    &parser.Node{parser.LITERAL, "Node34"},
	})
	openingTag, closingTag, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	expectedOpeningTag, expectedClosingTag = `<spdx:Snippet rdf:nodeID="Node34">`, "</spdx:Snippet>"
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if openingTag != expectedOpeningTag {
		t.Errorf("wrong opening tag. expected %s, got %s", expectedOpeningTag, openingTag)
	}
	if closingTag != expectedClosingTag {
		t.Errorf("wrong closing tag. expected %s, got %s", expectedClosingTag, closingTag)
	}

	// TestCase 5: more than one nodeID attribute (an invalid case)
	triples = append(triples, triples[1])
	_, _, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	if err == nil {
		t.Error("function should raise an error when there are more than one triples having rdf:nodeID predicate")
	}

	// resetting triples with only no rdf:nodeID and only one rdf:type attribute.
	triples = triples[:1]

	// TestCase 6: invalid rdf:type uri
	triples[0].Object = nodes[3]
	// predicate of triples[0] is rdf:type. It expects the object uri of
	// type baseName:fragment. but nodes[3] is a blank node with ID "N4"
	_, _, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
	if err == nil {
		t.Error("expected an invalid uri error")
	}

	// TestCase 7: Valid case where we have a rdf:about tag
	triples = []*parser.Triple{
		{ // rdf:type="http://spdx.org/rdf/terms#Snippet"
			Subject:   nodes[0],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
			Object:    &parser.Node{NodeType: parser.IRI, ID: "http://spdx.org/rdf/terms#Snippet"},
		},
		{ // rdf:about="http://spdx.org/rdf/terms#Snippet132"
			Subject:   nodes[0],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
			Object:    &parser.Node{NodeType: parser.IRI, ID: "http://spdx.org/rdf/terms#Snippet132"},
		},
	}
	_, _, err = getOpeningAndClosingTags(triples, rdfNSAbbrev, invSchemaDefinition, tab, nodes[0])
}

func Test_getRootTagFromSchemaDefinition(t *testing.T) {
	schemaDefinition := make(map[string]uri.URIRef)
	tab := "  "

	// TestCase 1: empty schema definition
	rootTag := getRootTagFromSchemaDefinition(schemaDefinition, tab)
	expectedOp := "<rdf:RDF>"
	if rootTag != expectedOp {
		t.Errorf("incorrect output. expected %s, found %s", expectedOp, rootTag)
	}

	// TestCase 2: only one url in the schemaDefinition
	rdfURI, _ := uri.NewURIRef(parser.RDFNS)
	schemaDefinition["rdf"] = rdfURI
	rootTag = getRootTagFromSchemaDefinition(schemaDefinition, tab)
	expectedOp = `<rdf:RDF
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`
	if rootTag != expectedOp {
		t.Errorf("incorrect output. expected %s, found %s", expectedOp, rootTag)
	}

	// Note: not checking for outputs with more than one uri because, the
	// input to the function is a map object and the function iterates over the keys and
	// linearly generates the output. Since the keys in a map are unordered,
	// the output might be different on every run.
}
