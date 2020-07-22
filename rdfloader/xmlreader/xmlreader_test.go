package rdfloader

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
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
	rand.Seed(time.Now().Unix()) // seed to ensure different pseudo-random number.
	file.name = fmt.Sprintf("sample_file_for_test%v.rdf", rand.Int())

	// writing the content to file and returning the err.
	err := ioutil.WriteFile(file.name, []byte(content), 777)
	return file, err
}

func (file *TestFile) Delete() {
	// delete the temporary file
	err := os.Remove(file.name)
	if err != nil {
		panic("couldn't delete test file " + file.name) // didnt' expect any error
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
	if err != nil {
		panic("cannot open test file")
	}
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
	xmlReader := xmlreaderFromString(SampleRDF)
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

	// TestCase 2: root block followed by any tag or chars must raise an error
	xmlReader = xmlreaderFromString(SampleRDF + "\n<tag/>")
	_, err = xmlReader.Read()
	if err == nil {
		t.Error("expected Read() to raise an error")
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
	// TestCase 1: prolog with only one block
	testString := `<? xml version="1.0" ?>
				   <rdf:RDF />`
	xmlReader := xmlreaderFromString(testString)
	block, err := xmlReader.readBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if block.OpeningTag.SchemaName != "rdf" || block.OpeningTag.Name != "RDF" {
		t.Errorf("expected <rdf:RDF>, found <%v:%v>", block.OpeningTag.SchemaName, block.OpeningTag.Name)
	}

	// TestCase 2: block with invalid prolog
	testString = `<? xml version="1.0" >`
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readBlock()
	if err == nil {
		t.Error("expected an error reporting invalid prolog")
	}

	// TestCase 3: complete block without a separate closing tag.
	testString = `<rdf:RDF/>`
	xmlReader = xmlreaderFromString(testString)
	block, _ = xmlReader.readBlock()
	if len(block.Children) != 0 {
		t.Errorf("expected block to have no children. Found %v children", len(block.Children))
	}
	if block.OpeningTag.SchemaName != "rdf" || block.OpeningTag.Name != "RDF" {
		t.Errorf("expected block opening tag to be <rdf:RDF>, found: <%v:%v>", block.OpeningTag.SchemaName, block.OpeningTag.Name)
	}

	// TestCase 4: different opening and closing tag
	testString = `<rdf:RDF> </rdf:rdf>`
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readBlock()
	if err == nil {
		t.Error("should've raised an error reporting different opening and closing tags")
	}

	// TestCase 5: valid case: block with single attribute
	xmlReader = xmlreaderFromString(SampleRDF)
	_, err = xmlReader.readBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// TestCase 6: invalid case: block with invalid cdata ( cdata without end tag )
	testString = `
<spdx:extractedText>
    <![CDATA[License by Nomos.
</spdx:extractedText>`
    xmlReader = xmlreaderFromString(testString)
    _, err = xmlReader.readBlock()
    if err == nil {
    	t.Error("expected an error saying eof reading cdata end tag")
	}

	// TestCase 7: valid case: block with valid cdata.
	testString = `
<spdx:extractedText>
    <![CDATA[License by Nomos.]]>
</spdx:extractedText>`
	xmlReader = xmlreaderFromString(testString)
	block, err = xmlReader.readBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedBlock := Block{
		OpeningTag: Tag{
			SchemaName: "spdx",
			Name:       "extractedText",
			Attrs:      nil,
		},
		Value:      "<![CDATA[License by Nomos.]]>",
		Children:   nil,
	}
	if !reflect.DeepEqual(block, expectedBlock) {
		t.Errorf("mismatching output. \nExpected: %+v. \nFound: %+v", expectedBlock, block)
	}
}

func TestXMLReader_readClosingTag(t *testing.T) {
	// TestCase 1: empty input, should raise an error
	testString := ""
	xmlReader := xmlreaderFromString(testString)
	_, err := xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("expected an error, found nil")
	}

	// TestCase 2: valid input with name and schemaName
	xmlReader = xmlreaderFromString("</rdf:RDF>")
	closingTag, err := xmlReader.readClosingTag()
	if err != nil {
		t.Errorf("unexpected errof: %v", err)
	}
	if closingTag.SchemaName != "rdf" || closingTag.Name != "RDF" {
		t.Errorf("invalid closing tag. Expected </rdf:RDF> found </%v:%v>", closingTag.SchemaName, closingTag.Name)
	}

	// TestCase 2: valid input without schemaName
	testString = `</tag>`
	xmlReader = xmlreaderFromString(testString)
	closingTag, err = xmlReader.readClosingTag()
	if err != nil {
		t.Errorf("unexpected errof: %v", err)
	}
	if closingTag.SchemaName != "" {
		t.Errorf("invalid closing tag. Expected </tag> found </%v:%v>", closingTag.SchemaName, closingTag.Name)
	}
	if closingTag.Name != "tag" {
		t.Errorf("invalid closing tag. Expected </tag> found </%v>", closingTag.Name)
	}

	// TestCase 3: invalid input with space before colon
	testString = "</rdf :RDF>"
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("reading invalid closing tag didn't raise any error")
	}

	// TestCase 4: invalid input with WHITESPACE after colon
	testString = `</rdf:
 						RDF>`
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("reading invalid closing tag didn't raise any error")
	}

	// TestCase 5: invalid input with stray characters before the closing tag end.
	testString = `</rdf:RDF stray-chars>`
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("reading invalid closing tag didn't raise any error")
	}

	// TestCase 6: invalid input which doesn't start with a </
	testString = `<rdf:RDF>`
	xmlReader = xmlreaderFromString(testString)
	_, err = xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("reading invalid closing tag didn't raise any error")
	}
}

func TestXMLReader_readColonPair(t *testing.T) {

	// Testcase 1: Valid input.
	testString := `rdf:RDF>`
	xmlReader := xmlreaderFromString(testString)
	pair, colonFound, err := xmlReader.readColonPair(1 << '>')
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if colonFound != true {
		t.Error("input had colon. program didn't find any colon")
	}
	if pair.First != "rdf" || pair.Second != "RDF" {
		t.Errorf("wrong pair found. Expected: rdf:RDF, found %v:%v", pair.First, pair.Second)
	}
	// after reading the pair, next character should be >
	nextRune, err := xmlReader.readARune()
	if err != nil {
		t.Errorf("unexpected error reading next character after reading attribute. Error: %v", err)
	}
	if nextRune != '>' {
		t.Errorf("next rune to be read should be >, found %v", nextRune)
	}

	// TestCase 2: valid input without colon:
	testString = `tag>`
	xmlReader = xmlreaderFromString(testString)
	pair, colonFound, err = xmlReader.readColonPair(1 << '>')
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if colonFound == true {
		t.Error("function found a colon when it doesn't exist in the input")
	}
	if pair.First != "tag" && pair.Second != "" {
		t.Errorf("expected pair: ('%v', '%v'), found pair: ('%v', '%v')", "tag", "", pair.First, pair.Second)
	}

	// TestCase 3: valid tag, invalid delimiter
	testString = `rdf:RDF>`
	xmlReader = xmlreaderFromString(testString)
	_, _, err = xmlReader.readColonPair(1 << ' ')
	if err == nil {
		t.Errorf("expected program to raise EOF error. error raised: %v", err)
	}
}

func TestXMLReader_readOpeningTag(t *testing.T) {
	// TestCase 1: empty input which doesn't have any opening tag.
	testString := ""
	xmlReader := xmlreaderFromString(testString)
	_, err := xmlReader.readClosingTag()
	if err == nil {
		t.Errorf("expected an EOF error, found nil")
	}

	// TestCase 2: invalid input which doesn't start with <
	testString = "extra chars <rdf:RDF>"
	xmlReader = xmlreaderFromString(testString)
	_, _, _, err = xmlReader.readOpeningTag()
	if err == nil {
		t.Errorf("expected an error, found nil")
	}

	// TestCase 3: invalid input which doesn't have < tag
	testString = "extra chars"
	xmlReader = xmlreaderFromString(testString)
	_, _, _, err = xmlReader.readOpeningTag()
	if err == nil {
		t.Errorf("expected an EOF error, found nil")
	}

	// TestCase 4: invalid input : closing tag.
	testString = "</tag>"
	xmlReader = xmlreaderFromString(testString)
	_, _, _, err = xmlReader.readOpeningTag()
	if err == nil {
		t.Error("expected an error, found nil")
	}

	// TestCase 5: valid prolog as an opening tag.
	testString = `<? xml version="1.0" ?>`
	xmlReader = xmlreaderFromString(testString)
	_, isProlog, _, err := xmlReader.readOpeningTag()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isProlog {
		t.Error("input has a prolog, program didn't found any")
	}

	// TestCase 6: invalid prolog as an opening tag.
	testString = `<? xml version="1.0" >`
	xmlReader = xmlreaderFromString(testString)
	_, isProlog, _, err = xmlReader.readOpeningTag()
	if err == nil {
		t.Error("expected an EOF error")
	}
	if !isProlog {
		t.Errorf("program didn't identify prolog tag")
	}

	// TestCase 7: input tag with only name and schema name.
	testString = `<rdf:RDF 
					>`
	xmlReader = xmlreaderFromString(testString)
	openingTag, isProlog, blockComplete, err := xmlReader.readOpeningTag()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isProlog {
		t.Error("program incorrectly identified opening tag as a prolog")
	}
	if blockComplete {
		t.Error("block was incorrectly tagged as complete")
	}
	if openingTag.SchemaName != "rdf" || openingTag.Name != "RDF" {
		t.Errorf("Expected tag: <rdf:RDF>, found: <%v:%v>", openingTag.SchemaName, openingTag.Name)
	}
	// there shouldn't be any attribute in an empty tag.
	if len(openingTag.Attrs) != 0 {
		t.Errorf("found unwanted attributes in an empty tag")
	}

	// TestCase 8: opening tag of type <rdf:RDF /> where block completed.
	testString = `< rdf:RDF
                           />`
	xmlReader = xmlreaderFromString(testString)
	openingTag, isProlog, blockComplete, err = xmlReader.readOpeningTag()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isProlog {
		t.Error("program incorrectly identified opening tag as a prolog")
	}
	if !blockComplete {
		t.Error("block didn't identify block as complete")
	}
	if openingTag.SchemaName != "rdf" || openingTag.Name != "RDF" {
		t.Errorf("Expected tag: <rdf:RDF>, found: <%v:%v>", openingTag.SchemaName, openingTag.Name)
	}
	// there shouldn't be any attributes in an empty tag.
	if len(openingTag.Attrs) != 0 {
		t.Error("found unwanted attributes in an empty tag")
	}

	// TestCase 9: opening tag with incomplete attribute definition.
	testString = `<rdf:RDF xmlns:rdf=>`
	xmlReader = xmlreaderFromString(testString)
	_, _, _, err = xmlReader.readOpeningTag()
	if err == nil {
		t.Error("input tag with incomplete attribute didn't raise any error")
	}

	// TestCase 10: valid input: opening tag with a single attribute.
	testString = `< rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" />`
	xmlReader = xmlreaderFromString(testString)
	openingTag, _, blockComplete, err = xmlReader.readOpeningTag()
	if !blockComplete {
		t.Error("a complete block wasn't tagged as complete")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if nAttrs := len(openingTag.Attrs); nAttrs != 1 {
		t.Errorf("expected only 1 attribute, found %v attributes.", nAttrs)
	}
}

func TestXMLReader_readCDATA(t *testing.T) {
	CDATA_OPENING := "<![CDATA["
	data := "random data"
	CDATA_CLOSING := "]]>"

	// TestCase 1: reading in an empty file must raise an error
	fileContent := ""
	xmlReader := xmlreaderFromString(fileContent)
	_, err := xmlReader.readCDATA()
	if err == nil {
		t.Errorf("expected an eof reading from an empty file")
	}

	// TestCase 2: file-pointer doesn't start with a CDATA Tag:
	// Must raise an error
	fileContent = "<rdf:RDF> </rdf:RDF>"
	xmlReader = xmlreaderFromString(fileContent)
	_, err = xmlReader.readCDATA()
	if err == nil {
		t.Errorf("expected an error saying not a valid cdata tag")
	}

	// TestCase 3: similar to TC2 but the file has a CDATA Tag which is prefixed by blank chars.
	// Must raise an error. as the function expects blank chars to be stripped before input
	fileContent = "  " + CDATA_OPENING + data + CDATA_CLOSING
	xmlReader = xmlreaderFromString(fileContent)
	_, err = xmlReader.readCDATA()
	if err == nil {
		t.Errorf("expected an error saying not a valid cdata tag")
	}

	// TestCase 4: CDATA tag without closing must raise an error
	fileContent = CDATA_OPENING + data
	xmlReader = xmlreaderFromString(fileContent)
	_, err = xmlReader.readCDATA()
	if err == nil {
		t.Errorf("expected an error saying eof reading closing tag")
	}

	// TestCase 5: Valid CDATA Tag
	fileContent = CDATA_OPENING + data + CDATA_CLOSING + " some other content.... "
	xmlReader = xmlreaderFromString(fileContent)
	output, err := xmlReader.readCDATA()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOutput := CDATA_OPENING + data + CDATA_CLOSING
	if output != expectedOutput {
		t.Errorf("expected: %s as the output, found: %s", fileContent, output)
	}
}
