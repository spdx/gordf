package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/RishabhBhatnagar/gordf/rdfloader/parser"
	xmlreader "github.com/RishabhBhatnagar/gordf/rdfloader/xmlreader"
	"github.com/RishabhBhatnagar/gordf/rdfwriter"
	"io"
	"strings"
)

func xmlreaderFromString(fileContent string) xmlreader.XMLReader {
	return xmlreader.XMLReaderFromFileObject(bufio.NewReader(io.Reader(bytes.NewReader([]byte(fileContent)))))
}

func main() {
	testString := `
		<?xml version="1.0"?>
		<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		    xmlns:dc="http://purl.org/dc/elements/1.1/">
		    <rdf:Description rdf:about="http://www.w3.org/">
			    <dc:title>World Wide Web Consortium</dc:title> 
		    </rdf:Description>
		</rdf:RDF>
	`

	// in the real world, this will be replaced with
	// xmlreader.XMLReaderFromFileObject call for getting a new file xmlreader
	xmlReader := xmlreaderFromString(testString)
	xmlReader, _ = xmlreader.XMLReaderFromFilePath("RDF Files/1.xml")
	// parsing the underlying xml structure of rdf file.
	rootBlock, _ := xmlReader.Read()

	// creating a new parser object
	rdfParser := parser.New()
	// sets rdf triples from the xml elements from the xmlreader
	rdfParser.Parse(rootBlock)

	// Example 1:
	// Getting string of all the triples
	tab := "    "
	opString, err := rdfwriter.TriplesToString(rdfParser.Triples, rdfParser.SchemaDefinition, tab)
	if err != nil {
		panic("error in a valid example")
	}
	asterisks := strings.Repeat("*", 33)
	fmt.Println(asterisks, "OUTPUT String", asterisks)
	fmt.Println(opString)

	// Example 2: writing rdf-triples to a file.
	var b bytes.Buffer

	// the output will be written to the buffer.
	rdfwriter.WriteToFile(&b, rdfParser.Triples, rdfParser.SchemaDefinition, tab)
}
