package main

import (
	"fmt"
	rdfloader "github.com/spdx/gordf/rdfloader"
	"os"
)

func main() {
	// expects user to enter the file name.
	// sample run :
	// 		go run exampleRDFLoader.go input.rdf

	// checking if input arguments are ok.
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <input.rdf>\n", os.Args[0])
		fmt.Printf("\tTo parse the <input.rdf> file and\n")
		fmt.Printf("\tPrint some of it's Triples\n")
		os.Exit(1) // there was an error processing input.
	}

	filePath := os.Args[1]

	rdfParser, err := rdfloader.LoadFromFilePath(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// maximum number of triples to display.
	maxNTriples := 10
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
