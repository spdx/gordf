package rdfloader

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func xmlreaderFromString(fileContent string) XMLReader {
	return XMLReaderFromFileObject(bufio.NewReader(io.Reader(bytes.NewReader([]byte(fileContent)))))
}

func TestXMLReader_ignoreWhiteSpace(t *testing.T) {
	// string starting with 4 blank characters.
	fileContent1 := "\n \r\tsample string"

	// string without any blank space at the start.
	fileContent2 := "sample string"

	// new reader with above string which starts with blank chars.
	xmlReader1 := xmlreaderFromString(fileContent1)
	// reader for string without any blank char at start
	xmlReader2 := xmlreaderFromString(fileContent2)

	// Test Case 1
	nWS, err := xmlReader1.ignoreWhiteSpace()
	if err != nil {
		t.Error(err)
	}
	if nWS != 4 {
		t.Errorf("ignored %v spaces, expected it to ignore %v spaces", nWS, 4)
	}

	// Test Case 2
	nWS, err = xmlReader2.ignoreWhiteSpace()
	if err != nil {
		t.Error(err)
	}
	if nWS != 0 {
		t.Errorf("ignored %v spaces, expected it to ignore %v spaces", nWS, 0)
	}
}

func TestXMLReader_peekARune(t *testing.T) {
	testCases := []string{
		"string - alphabets",
		"123456 - numbers",
		"@#$%%^ - non-alphanumeric chars",
	}

	for _, testCase := range testCases {
		xmlReader := xmlreaderFromString(testCase)
		r, err := xmlReader.peekARune()
		if err != nil {
			t.Error(err)
		} else if expected := rune(testCase[0]); r != expected {
			t.Errorf("Expected %v, Found %v", expected, r)
		}
	}

	// testing if the function returns an error upon eof
	xmlReader := xmlreaderFromString("")
	_, err := xmlReader.peekARune()
	if err == nil {
		t.Error("expected function to raise an error")
	}
}

func TestXMLReader_peekNBytes(t *testing.T) {
	type TestCase struct {
		input       string
		numberChars int
		output      []byte
		err         error
	}
	testCases := []TestCase{
		TestCase{"string", 3, []byte("str"), nil},
		TestCase{"string", 6, []byte("string"), nil},
		TestCase{"string", 0, []byte{}, nil},
		TestCase{"string", 7, []byte("string"), io.EOF},
	}
	for _, testCase := range testCases {
		xmlReader := xmlreaderFromString(testCase.input)
		nBytes, err := xmlReader.peekNBytes(testCase.numberChars)
		if bytes.Compare(nBytes, testCase.output) != 0 {
			t.Errorf("Expected %v, Found %v", string(nBytes), string(testCase.output))
		}
		if err != testCase.err {
			t.Errorf("expected to raise %v error", testCase.err)
		}
	}
}

func TestXMLReader_readARune(t *testing.T) {
	type TestCase struct {
		input       string
		numberChars int
		output      []byte
		err         error
	}
	testCases := []TestCase{
		TestCase{"string", 3, []byte("str"), nil},
		TestCase{"string", 6, []byte("string"), nil},
		TestCase{"string", 0, []byte{}, nil},
		TestCase{"string", 7, []byte("string"), io.EOF},
	}
	for _, testCase := range testCases {
		xmlReader := xmlreaderFromString(testCase.input)
		nBytes, err := xmlReader.peekNBytes(testCase.numberChars)
		if bytes.Compare(nBytes, testCase.output) != 0 {
			t.Errorf("Expected %v, Found %v", string(nBytes), string(testCase.output))
		}
		if err != testCase.err {
			t.Errorf("expected to raise %v error", testCase.err)
		}
	}
}

func TestXMLReader_readNBytes(t *testing.T) {
	xmlReader := xmlreaderFromString("string")
	op, err := xmlReader.readNBytes(4)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if bytes.Compare(op, []byte("stri")) != 0 {
		t.Errorf("Expected: %v, Found: %v", "stri", string(op))
	}
}

func TestXMLReader_readTill(t *testing.T) {
	xmlReader := xmlreaderFromString("hello world")
	op, err := xmlReader.readTill(1 << ' ')
	if string(op) != "hello" {
		t.Errorf("Expected %v, Found: %v", "hello", string(op))
	}

	// next character must be a space.
	// That is, delimiter mustn't be consumed while reading.
	r, err := xmlReader.readARune()
	if err != nil {
		t.Error(err)
	}
	if r != ' ' {
		t.Errorf("Expected %v, Found %v", ' ', r)
	}
}

func TestXMLReader_readTillString(t *testing.T) {
	// TestCase 1: searching in an empty file must raise an eof error
	xmlReader := xmlreaderFromString("")
	_, err := xmlReader.readTillString("any string")
	if err == nil {
		t.Error("expected an eof error")
	}

	// TestCase 2: searching in a file with delimiter at the end.
	// it shouldn't raise any error.
	delimiter := "example"
	fileContent := delimiter
	xmlReader = xmlreaderFromString(fileContent)
	output, err := xmlReader.readTillString(delimiter)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if len(output) > 0 {
		t.Error("expected output to be empty as the file didn't had any chars before the delimiter")
	}

	// TestCase 3: delimiter not in the file must return an empty output with an eof error.
	fileContent = "sample content"
	delimiter = "delim"
	xmlReader = xmlreaderFromString(fileContent)
	output, err = xmlReader.readTillString(delimiter)
	if err == nil {
		t.Error("expected an eof error finding delimiter")
	}
	if len(output) > 0 {
		t.Errorf("output should be empty when the delimiter is not found which ran into an error. found %v", output)
	}

	// TestCase 4: Valid case: delimiter present in the string.
	delimiter = "delimiter"
	fileContent = "some random chars" + delimiter
	xmlReader = xmlreaderFromString(fileContent)
	output, _ = xmlReader.readTillString(delimiter)
	expected := strings.Split(fileContent, delimiter)[0]
	if string(output) != expected {
		t.Errorf("faulty parsing. expected: %s, found: %s", expected, string(output))
	}
}
