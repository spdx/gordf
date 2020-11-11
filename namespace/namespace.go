package namespace

/*
Provides a Namespace class.
USAGE:
	>>> ns, _ := namespace.New("https://spdx.org/rdf/terms")
    >>> ns.Get("d4e2952")
    {https://spdx.org/rdf/terms#d4e2952}
	>>> ns.Get("Tag")
	{https://spdx.org/rdf/terms#Tag}

	// Alternative initializations of namespace without abstraction
	>> ns, _ := namespace.Namespace{uri.URIRef{"https://spdx.org/rdf/terms"}}
*/

import (
	"github.com/spdx/gordf/uri"
)

type Namespace struct {
	base uri.URIRef
}

// Provides an abstraction to create a Namespace directly from uri string.
func New(namespace string) (ns Namespace, err error) {
	uriref, err := uri.NewURIRef(namespace)
	if err != nil {
		return
	}
	return Namespace{uriref}, nil
}

// Appends a fragment string at the end of the namespace string.
func (ns *Namespace) Get(fragment string) uri.URIRef {
	return ns.base.AddFragment(fragment)
}
