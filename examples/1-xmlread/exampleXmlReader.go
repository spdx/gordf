// SPDX-License-Identifier: MIT License

// referred the documentation from
// https://github.com/spdx/tools-golang/blob/master/examples/1-load/example_load.go

// Example for: *xmlreader*

// This example demonstrates loading an xml file from the disk into the
// memory and printing some of the tags to validate it's correctness.

package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	reader "github.com/spdx/gordf/rdfloader/xmlreader"
)

func printBlockHeader(block reader.Block) {
	// Prints name and attributes of the block.
	fmt.Printf("Name: <%v:%v>\n", block.OpeningTag.SchemaName, block.OpeningTag.Name)

	fmt.Printf("Tag Attributes:\n")
	for _, attr := range block.OpeningTag.Attrs {
		// schemaName for an attribute is optional
		if attr.SchemaName == "" {
			fmt.Printf("\t%v=\"%v\"\n", attr.Name, attr.Value)
		} else {
			fmt.Printf("\t%v:%v=\"%v\"\n", attr.SchemaName, attr.Name, attr.Value)
		}
	}
}

func main() {
	// expects user to enter the file path along with it's name as an argument
	// while running the file

	// Sample run:
	// go run exampleXmlReader.go ../sample-docs/rdf/input.rdf

	// check if we've received the right number of arguments
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <input.rdf>\n", os.Args[0])
		fmt.Printf("\tLoad the <input.rdf> file into memory and\n")
		fmt.Printf("\tPrint some of it's tags.")
		os.Exit(1) // there was an error processing input.
	}

	// filePath indicates path to the file along with the filename.
	filePath := os.Args[1]

	// in xmlReader, we've two options of reading the file.
	// 		1. using the bufio file object
	// 		2. using the file path

	// Method 1: using file objects
	fileHandler, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening %v: %v\n", filePath, err)
		os.Exit(1)
	}
	xmlReader1 := reader.XMLReaderFromFileObject(bufio.NewReader(fileHandler))
	root1, err := xmlReader1.Read()
	if err != nil {
		fmt.Printf("error while parsing %v: %v\n", filePath, err)
		os.Exit(1)
	}
	fileHandler.Close()

	// Method 2: using file paths
	xmlReader2, err := reader.XMLReaderFromFilePath(filePath)
	if err != nil {
		fmt.Printf("error opening %v: %v", filePath, err)
		os.Exit(1)
	}
	root2, err := xmlReader2.Read()
	if err != nil {
		fmt.Printf("error while parsing %v: %v\n", filePath, err)
		os.Exit(1)
	}

	// comparing the results from both the reader instances.
	if !reflect.DeepEqual(root1, root2) {
		fmt.Println("outputs from both approach is not same.")
		fmt.Println("Something wrong with the implementation")
		os.Exit(1)
	}

	// printing the root tag
	fmt.Println(strings.Repeat("#", 80))
	fmt.Println("Root Tag:")
	printBlockHeader(root1)
	fmt.Println(strings.Repeat("#", 80))

	// maximum number of children to print
	maxNChild := 10

	// iterating over the children of the root tag.
	for i, childBlock := range root1.Children {
		if i >= maxNChild {
			break
		}
		fmt.Println()
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Child %v:\n", i+1)
		printBlockHeader(*childBlock)
		fmt.Println(strings.Repeat("=", 80))
	}
}
