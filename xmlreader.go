package rdfloader


/**
 * This module provides the functions needed to read a file tag by tag.
 * Since the documents are written in rdf/xml,
 * Creating an xml reader for reading rdf tags.
 */

import (
	"bufio"
	"errors"
	"io"
	"os"
	"unicode"
)


const WHITESPACE = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '


type XMLReader struct {
	fileReader *bufio.Reader
}


type Attribute struct {
	Name string
	SchemaName string
	Value string
}


type pair struct {
	first interface{}
	second interface{}
}


type Tag struct {
	SchemaName string
	Name string
	Attrs []Attribute
}


type Block struct {
	// A block is a valid sub-xml.
	// for example:
	// 		1. <tag />
	// 		2. <tag attr="attr" />
	//      3. <tag> value </tag>
	//      4. <parent> <child> value </child> </parent>
	OpeningTag Tag
	Value string
	Children []*Block
}


func (xmlReader *XMLReader) peekARune() (r rune, err error) {
	// returns next character in the file without affecting the file pointer
	singleByte, err := xmlReader.fileReader.Peek(1)
	return rune(singleByte[0]), err
}


func (xmlReader *XMLReader) readARune() (rune, error) {
	singleByteArray := make([]byte, 1)
	_, err := xmlReader.fileReader.Read(singleByteArray)
	return rune(singleByteArray[0]), err
}


func (xmlReader *XMLReader) readTill(delim uint64) ([]rune, error) {
	// reads the input file rune by rune till the target rune is found
	//		or eof is reached.
	// Note: it doesn't include the target rune in the read word.
	var buffer []rune
	for {
		r, err := xmlReader.fileReader.Peek(1)
		if err == nil {
			// checking if the read rune is same as any of the delimiters' mask
			if (delim & (1 << r[0])) != 0 {
				// current char is same as one of the delimiters.
				return buffer, nil
			}

			// moving file pointer one character ahead.
			xmlReader.readARune()

			// current character is not one of the delimiters.
			buffer = append(buffer, rune(r[0]))
		} else {
			return buffer, err
		}
	}
}


func (xmlReader *XMLReader) peekNBytes(n int) ([]byte, error) {
	return xmlReader.fileReader.Peek(n)
}


func (xmlReader *XMLReader) readNBytes(n int) (nextNBytes []byte, err error) {
	next2Bytes := make([]byte, n)
	_, err = xmlReader.fileReader.Read(next2Bytes)
	return next2Bytes, err
}


func (xmlReader *XMLReader) ignoreWhiteSpace() (nWS int, err error) {
	// advance the file pointer until a non-blank character is found.
	// nWS: number of whitespaces which were stripped.
	for {
		char, err := xmlReader.peekARune()
		if err != nil || !unicode.IsSpace(char) {
			return nWS, err
		}
		nWS++
		xmlReader.readARune()
	}
}


func (xmlReader *XMLReader) readColonPair(delim uint64) (pair pair, colonFound bool, err error) {
	// reads a:b into a Pair Object.
	word, err := xmlReader.readTill(delim)
	if err != nil { return }

	for i, r := range word {
		if r == ':' {
			colonFound = true
			pair.first = string(word[:i])
			latter := string(word[i+1:])
			if len(latter) == 0 {
				err = errors.New("expected a word after colon")
				return
			}
			pair.second = latter
			break
		}
	}
	if !colonFound {
		// no colon was found.
		pair.first = string(word)
	}
	return
}


func (xmlReader *XMLReader) readAttribute() (attr Attribute, err error) {
	// assumes the file pointer is pointing to the attribute name
	pair, colonExists, err := xmlReader.readColonPair(WHITESPACE | 1 << '=')
	if err != nil { return attr, err }
	if colonExists {
		attr.SchemaName = pair.first.(string)
		attr.Name = pair.second.(string)
	} else {
		attr.Name = pair.first.(string)
	}
	_, err = xmlReader.ignoreWhiteSpace()
	if err != nil { return attr, err }

	nextRune, err := xmlReader.peekARune()
	if err != nil { return attr, err }
	if nextRune != '=' {
		err = errors.New("expected an assignment sign (=)")
	}
	xmlReader.readARune()

	firstQuote, err := xmlReader.readARune()
	if firstQuote != '\'' && firstQuote != '"' {
		err = errors.New("assignment operator must be followed by an attribute enclosed within quotes")
	}

	// read till next quote.
	word, err := xmlReader.readTill(WHITESPACE | 1 << byte(firstQuote))
	if err != nil { return attr, err }

	secondQuote, _ := xmlReader.readARune()
	if firstQuote != secondQuote {
		return attr, errors.New("unexpected blank char. expected a closing quote")
	}

	attr.Value = string(word)
	return attr, err
}


func (xmlReader *XMLReader) readOpeningTag() (tag Tag, blockComplete bool, err error) {
	// Opening Tag can be:
	//		<tag[:schema]
	//			[attr=attr_val]
	//			[attr=attr_val]...	>
	// or
	//		<tag[:schema]
	//			[attr=attr_val]
	//			[attr=attr_val]...	/>
	// Second example is a completed block where no value or internal nodes were found.

	var word []rune

	// forward file pointer until a not-space character is found.
	// removing all blank characters before opening bracket.
	_, err = xmlReader.ignoreWhiteSpace()
	if err != nil { return }

	// find the opening angular bracket.
	// after stripping all the spaces, the next character to be read should be '<'
	//   If the next character is not '<',
	//       there are few chars before opening tag. Which is not allowed.
	word, err = xmlReader.readTill(1 << '<')
	if err == io.EOF {
		// we reached the end of the file while searching for a new tag.
		if len(word) > 0 {
			return tag, blockComplete, errors.New("found stray characters at EOF")
		} else {
			// no new tags were found.
			return tag, blockComplete, io.EOF
		}
	}
	if len(word) != 0 {
		return tag, blockComplete, errors.New("found extra chars before tag start")
	}

	// next char is an opening angular bracket.
	xmlReader.readARune()
	xmlReader.ignoreWhiteSpace() // there shouldn't be any for a well-formed rdf/xml document.

	nextRune, _ := xmlReader.peekARune()
	if nextRune == '/' {
		return tag, blockComplete, errors.New("unexpected closing tag")
	}

	// reading the next word till we reach a colon or a blank-char or a closing angular bracket.
	pair, colonExist, err := xmlReader.readColonPair(1 << '>' | WHITESPACE | 1 << '/')
	if err != nil { return }

	if colonExist {
		tag.SchemaName = pair.first.(string)
		tag.Name = pair.second.(string)
	} else {
		tag.Name = pair.first.(string)
	}

	delim, _ := xmlReader.readARune() // read the delimiter.

	switch delim {
	case '>':
		// found end of tag. entire tag was parsed.
		return

	case '/':
		// "<[schemaName:]tag /" was parsed. expecting next character to be a closing angular bracket.
		blockComplete = true

		nextRune, err := xmlReader.peekARune()
		if err != nil {
			return tag, blockComplete, err
		}

		if nextRune == '>' {
			xmlReader.readARune() // advancing file pointer by 1.
		} else {
			err = errors.New("expected closing angular bracket after /")
		}
		return tag, blockComplete, err
	}

	// "<[schemaName:]tagName WhiteSpace" is parsed till now.

	_, err = xmlReader.ignoreWhiteSpace()
	if err != nil { return }

	nextRune, err = xmlReader.peekARune()
	if err != nil { return }

	if nextRune == '>' {
		// opening tag didn't had any attributes.
		tag.Name = string(word)
		return
	}

	// there are some attributes to be read.
	for !(nextRune == '>' || nextRune == '/') {
		attr, err := xmlReader.readAttribute()
		if err != nil { return tag, blockComplete, err }

		tag.Attrs = append(tag.Attrs, attr)
		_, err = xmlReader.ignoreWhiteSpace()
		if err != nil { return tag, blockComplete, err }

		nextRune, err = xmlReader.peekARune()
		if err != nil { return tag, blockComplete, err }
	}

	nextRune, _ = xmlReader.readARune()

	if nextRune == '/' {
		// "<[schemaName:]tag /" was parsed. expecting next character to be a closing angular bracket.
		blockComplete = true

		nextRune, err := xmlReader.peekARune()
		if err != nil { return tag, blockComplete, err }

		if nextRune == '>' {
			xmlReader.readARune() // advancing file pointer by 1.
		} else {
			err = errors.New("expected closing angular bracket after /")
		}
	}
	return tag, blockComplete, err
}


func (xmlReader *XMLReader) readClosingTag() (closingTag Tag, err error) {
	// expects white space to be stripped before receiving
	next2Bytes, err := xmlReader.readNBytes(2)
	if err != nil {
		return closingTag, err
	}
	if string(next2Bytes) != "</" {
		return closingTag, errors.New("expected a closing tag")
	}

	pair, colonExists, err := xmlReader.readColonPair(1 << '>' | WHITESPACE)
	if err != nil {
		return closingTag, err
	}

	if colonExists {
		closingTag.SchemaName = pair.first.(string)
		closingTag.Name = pair.second.(string)
	} else {
		closingTag.Name = pair.first.(string)
	}

	xmlReader.ignoreWhiteSpace()
	nextChar, err := xmlReader.readARune()
	if err != nil {
		return closingTag, err
	}

	if nextChar != '>' {
		return closingTag, errors.New("expected a > char")
	}

	return closingTag, err
}


func (xmlReader *XMLReader) readBlock() (block Block, err error) {
	openingTag, blockComplete, err := xmlReader.readOpeningTag()
	if err != nil {
		return block, err
	}
	block.OpeningTag = openingTag

	if blockComplete {
		// tag was of this type: <schemaName:tagName />
		return block, err
	}

	xmlReader.ignoreWhiteSpace()
	// <schemaName:tagName> is read till now.
	nextRune, err := xmlReader.peekARune()
	if err != nil { return block, err }

	if nextRune != '<' {
		//	the tag must be wrapping a string resource within it.
		// tag is of type <schemaName:tagName> value </schemaName:tagName>
		word, err := xmlReader.readTill(1 << '<') // according to example, word = value.
		if err != nil {
			return block, err
		}
		block.Value = string(word)
	} else {
		// expecting a new tag or the closing tag of the currently read tag.
		nextTwoBytes, err := xmlReader.peekNBytes(2)
		if err != nil {
			return block, err
		}
		for string(nextTwoBytes) != "</" {
			// a new tag is found.
			childBlock, err := xmlReader.readBlock()
			if err != nil {
				return block, err
			}

			block.Children = append(block.Children, &childBlock)

			xmlReader.ignoreWhiteSpace()
			nextTwoBytes, err = xmlReader.peekNBytes(2)
			if err != nil {
				return block, err
			}
		}
	}

	closingTag, err := xmlReader.readClosingTag()
	if err != nil {
		return block, err
	}

	if closingTag.Name != closingTag.Name || closingTag.SchemaName != closingTag.SchemaName {
		// opening and closing tags are not equal.
		return block, errors.New("opening and closing tags doesn't match")
	}
	return block, err
}


func (xmlReader *XMLReader) Read() (rootBlock Block, err error) {
	rootBlock, err = xmlReader.readBlock()
	return rootBlock, err
}


func XMLReaderFromFileObject(fileObject *bufio.Reader) XMLReader {
	return XMLReader{fileObject}
}


func XMLReaderFromFilePath(filePath string) (xmlReader XMLReader, err error) {
	fileReader, err := os.Open(filePath)
	defer fileReader.Close()
	if err != nil {
		return xmlReader, err
	}
	return XMLReaderFromFileObject(bufio.NewReader(fileReader)), nil
}
