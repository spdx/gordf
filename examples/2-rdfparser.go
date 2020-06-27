package main

import (
	"fmt"
	"github.com/RishabhBhatnagar/gordf/rdfloader/parser"
	"os"
)

func main() {
	// expects user to enter the file name.
	// sample run :
	// 		go run 2-rdfparser.go input.rdf

	// checking if input arguments are ok.
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <input.rdf>\n", os.Args[0])
		fmt.Printf("\tTo parse the <input.rdf> file and\n")
		fmt.Printf("\tPrint some of it's Triples")
		os.Exit(1)  // there was an error processing input.
	}

	filePath := os.Args[1]
	rdfParser := parser.New()
	err := rdfParser.Parse(filePath)
	if err != nil {
		fmt.Printf("error parsing file: %v\n", err)
	}

	maxNTriples := 10
	if len(rdfParser.Triples) < maxNTriples {
		// in case the number of triples is less than the declared value.
		maxNTriples = len(rdfParser.Triples)
	}
	i := 0
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
