package main

import (
	"fmt"
	"github.com/spdx/gordf/rdfloader/parser"
	xmlreader "github.com/spdx/gordf/rdfloader/xmlreader"
	"os"
)

func main() {
	// expects user to enter the file name.
	// sample run :
	// 		go run exampleRDFParser.go input.rdf

	// checking if input arguments are ok.
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <input.rdf>\n", os.Args[0])
		fmt.Printf("\tTo parse the <input.rdf> file and\n")
		fmt.Printf("\tPrint some of it's Triples\n")
		os.Exit(1) // there was an error processing input.
	}

	filePath := os.Args[1]
	xmlReader, err := xmlreader.XMLReaderFromFilePath(filePath)
	if err != nil {
		// error reading the rdf file.
		fmt.Println(err.Error())
		os.Exit(1)
	}
	rootBlock, err := xmlReader.Read()
	if err != nil {
		// error parsing the xml content
		fmt.Println("error parsing the xml content of the rdf file")
	}
	rdfParser := parser.New()
	err = rdfParser.Parse(rootBlock)
	if err != nil {
		fmt.Printf("error parsing file: %v\n", err)
	}

	// the max number of triples to display
	maxNTriples := 10
	if len(rdfParser.Triples) < maxNTriples {
		// in case the number of triples is less than the declared value.
		maxNTriples = len(rdfParser.Triples)
	}

	i := 0
	// parser.Triples is a dictionary of the form {triple-hash => triple}
	for tripleHash := range rdfParser.Triples {
		if i == maxNTriples {
			break
		}
		i++
		triple := rdfParser.Triples[tripleHash]
		fmt.Printf("Triple %v:\n", i)
		fmt.Println("\tSubject:", triple.Subject)
		fmt.Println("\tPredicate:", triple.Predicate)
		fmt.Println("\tObject:", triple.Object)
	}
}
