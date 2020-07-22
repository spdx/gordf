package uri

import (
	"testing"
)

func TestNewURIRef(t *testing.T) {
	// TestCase 1: input is an empty string
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

	// TestCase 1: Invalid Fragment must return an empty uri
	fragment := "%%%%"
	newUri := uriref.AddFragment(fragment)
	if newUri.String() != "" {
		t.Errorf("invalid fragment must result in an empty uri. Found %v", newUri.String())
	}

	// TestCase 2: valid fragment
	fragment = "someFrag"
	newUri = uriref.AddFragment(fragment)
	expectedURI := uriString + "#" + fragment
	if newUri.String() != expectedURI {
		t.Errorf("expected: %v, found: %v", expectedURI, newUri)
	}

	// TestCase 3: valid fragmnet starting with a hash char.
	fragment = "#" + fragment
	newUri = uriref.AddFragment(fragment)
	if newUri.String() != expectedURI {
		t.Errorf("expected %v, found %v", expectedURI, newUri)
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
