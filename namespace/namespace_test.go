package namespace

import (
	"testing"
)

func TestNew(t *testing.T) {
	sampleURI := "https://www.spdx.org/rdf/terms"
	_, err := New(sampleURI)
	if err != nil {
		t.Errorf("error parsing a valid URI")
	}
}

func TestNamespace_Get(t *testing.T) {
	sampleURI := "https://www.spdx.org/rdf/terms"
	sampleNS, err := New(sampleURI)
	if err != nil {
		t.Errorf("error parsing a valid URI")
		return
	}
	fragment := "name"
	indexedURI := sampleNS.Get(fragment)
	if indexedURI.String() != sampleURI+"#"+fragment {
		t.Errorf("error adding fragment to the base URI")
	}
}
