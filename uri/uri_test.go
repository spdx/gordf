package uri

import (
	"testing"
)

func TestNewURIRef(t *testing.T) {
	// case when the input is an empty string
	_, err := NewURIRef("")
	if err == nil {
		t.Errorf("blank uri should've raised an error")
	}

	// case when the input is not empty and is an invalid uri.
	_, err = NewURIRef("%%")
	if err == nil {
		t.Errorf("expected function to raise an error for invalid input uri")
	}

	// case when the uri is valid and ends with a #
	uri := "https://www.spdx.org/rdf/terms#"
	uriref, err := NewURIRef(uri)
	if err != nil {
		t.Errorf(err.Error())
	}
	if uriref.String() != uri {
		t.Errorf("expected %v, found %v", uri, uriref)
	}

	// case when the uri is valid and doesn't end with a hash
	uri = "https://www.spdx.org/rdf/terms"
	uriref, err = NewURIRef(uri)
	if err != nil {
		t.Errorf(err.Error())
	}
	if exp := uri + "#"; uriref.String() != exp {
		t.Errorf("expected %v, found %v", exp, uriref)
	}
}

func TestURIRef_AddFragment(t *testing.T) {
	uriString := "https://www.someuri.com/valid/uri"
	uriref, _ := NewURIRef(uriString)

	// nothing much to test
	fragment := "someFrag"
	newUri := uriref.AddFragment(fragment)
	if exp := uriString + "#" + fragment; newUri.String() != exp {
		t.Errorf("expected: %v, found: %v", exp, newUri)
	}

}

func TestURIRef_String(t *testing.T) {
	// again, nothing much to test
	uriString := "https://www.someuri.com/valid/uri"
	uriref, _ := NewURIRef(uriString)
	if uriref.String() != uriString+"#" {
		t.Errorf("expected: %v, found: %v", uriString+"#", uriref.String())
	}
}
