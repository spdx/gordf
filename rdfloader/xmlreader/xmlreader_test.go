package rdfloader

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)



const SampleRDF = `
<rdf:RDF
    xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
    xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
	xmlns:custom="http://www.example.com/sample#">
	<rdf:Description rdf:about="https://www.other_domain/another_sample">
		<custom:Title>First Tag</custom:Title>
		<custom:Content>Some
						Multiline
						Content
		</custom:Content>
		<custom:BlankTag></custom:BlankTag>
		<custom:END custom:value="https://www.end.com/end_tag" />
	</rdf:Description>
</rdf:RDF>
`


type TestFile struct {
	name string
}

func InitTestFile(content string) (TestFile, error) {
	// creates a temporary file with given content in the current directory
	// and returns the pointer to testfile struct.

	file := TestFile{}
	rand.Seed(time.Now().Unix())  // seed to ensure different pseudo-random number.
	file.name = fmt.Sprintf("sample_file_for_test%v.rdf", rand.Int())

	// writing the content to file and returning the err.
	err := ioutil.WriteFile(file.name, []byte(content), 777)
	return file, err
}

func (file *TestFile) Delete() {
	// delete the temporary file
	err := os.Remove(file.name)
	if err != nil {
		panic("couldn't delete test file " + file.name )  // didnt' expect any error
	}
}

func TestXMLReaderFromFileObject(t *testing.T) {
	testFile, err := InitTestFile(SampleRDF)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer testFile.Delete()

	// nothing much to test here.
	// just creating a xmlReader.
	fh, err := os.Open(testFile.name)
	if err != nil { panic("cannot open test file") }
	defer fh.Close()
	XMLReaderFromFileObject(bufio.NewReader(fh))
}

func TestXMLReaderFromFilePath(t *testing.T) {
	// testing if the program raises error when the given file is not present.
	randomFileName := "@#$%^&*()_+"
	_, err := XMLReaderFromFilePath(randomFileName)
	if err == nil {
		t.Errorf("Expected function to raise error. Why is there a file named %v?", randomFileName)
	}

	// xmlreader when the file exists.
	testFile, err := InitTestFile(SampleRDF)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer testFile.Delete()

	xmlReader, err := XMLReaderFromFilePath(testFile.name)
	defer xmlReader.CloseFileObj()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestXMLReader_Read(t *testing.T) {
	testFile, err := InitTestFile(SampleRDF)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer testFile.Delete()

	fileObj, _ := os.Open(testFile.name)
	defer fileObj.Close()
	xmlReader := XMLReaderFromFileObject(bufio.NewReader(fileObj))
	root, err := xmlReader.Read()
	if err != nil {
		t.Errorf("unexpected error on a valid rdf file. %v", err)
	}

	// for the given SampleRDF, the root block has tags as children.
	// Value attribute of the root tag must be empty.
	if root.Value != "" {
		t.Errorf("expected root block value to be empty. Found %v", root.Value)
	}

	// root has only one child
	if nChildren := len(root.Children); nChildren != 1 {
		t.Errorf("expected root tag to have %v children, found %v children", 1, nChildren)
	}

	// root has 3 attributes
	if lenAttr := len(root.OpeningTag.Attrs); lenAttr != 3 {
		t.Errorf("expected root attribute to have %v attributes, found %v attributes", 3, lenAttr)
	}
}

func TestXMLReader_readAttribute(t *testing.T) {
	// the readAttribute assumes that the file pointer points to the name of the attribute.

	// string of rdf tag where the tag is having children.
	// for example: <tag attributes> <child> </child> </tag>
	tagIncompleteString := `xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
		xmlns:custom="http://www.example.com/sample#">
		... other tags ...
	`
	xmlReader := xmlreaderFromString(tagIncompleteString)
	attr, err := xmlReader.readAttribute()
	if err != nil {
		t.Errorf("unexpected error, %v", err)
	}

	// attribute has 3 properties: Name, SchemaName and Value.
	if attr.Name != "rdf" {
		t.Errorf("Wrong attribute name. Found %v, Expected %v", attr.Name, "rdf")
	}
	if attr.SchemaName != "xmlns" {
		t.Errorf("Wrong attribute schemaname. Found %v, Expected %v", attr.SchemaName, "xmlns")
	}
	if attr.Value != "http://www.w3.org/1999/02/22-rdf-syntax-ns#" {
		t.Errorf(
			"Wrong attribute value. Found %v, Expected %v",
			attr.Value,
			"http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		)
	}
}

func TestXMLReader_readBlock(t *testing.T) {
}

func TestXMLReader_readClosingTag(t *testing.T){
}

func TestXMLReader_readColonPair(t *testing.T) {
}

func TestXMLReader_readOpeningTag(t *testing.T) {
}