package rdfwriter

import (
	"github.com/RishabhBhatnagar/gordf/rdfloader/parser"
	"github.com/RishabhBhatnagar/gordf/uri"
	"reflect"
	"testing"
)

func max(n1, n2 int) int {
	if n1 > n2 {
		return n1
	}
	return n2
}

func getInvSchema() map[string]string {
	return map[string]string{
		"http://spdx.org/rdf/terms":                  "spdx",
		"http://www.w3.org/1999/02/22-rdf-syntax-ns": "rdf",
		"http://www.w3.org/2000/01/rdf-schema":       "rdfs",
	}
}

// returns a slice of n blank nodes.
// n > 0
func getNBlankNodes(n int) (blankNodes []*parser.Node) {
	blankNodes = make([]*parser.Node, max(0, n))
	// first blank nodes start with N1
	blankNodeGetter := parser.BlankNodeGetter{}
	for i := 0; i < n; i++ {
		newBlankNode := blankNodeGetter.Get()
		blankNodes[i] = &newBlankNode
	}
	return
}

func Test_getAdjacencyList(t *testing.T) {
	// TestCase 1
	// empty list of triples should have empty map as an output
	adjList := GetAdjacencyList([]*parser.Triple{})
	recoveryDS := GetNodeToTriples([]*parser.Triple{})
	if len(adjList) > 0 || len(recoveryDS) > 0 {
		t.Errorf("empty input is having non-empty output. Adjacency List: %v, recoveryDS: %v", adjList, recoveryDS)
	}

	// TestCase 2
	// modelling a simple graph depicted as follows:
	//              (N1)
	//       (N0) ---------> (N2)
	//        |
	//   (N3) |
	//        |
	//        v
	//       (N4)
	//
	// Triples for the above graph will be:
	//   1. N0 -> N1 -> N2
	//   2. N0 -> N3 -> N4
	blankNodes := getNBlankNodes(5)
	triples := []*parser.Triple{
		{
			Subject:   blankNodes[0],
			Predicate: blankNodes[1],
			Object:    blankNodes[2],
		}, {
			Subject:   blankNodes[0],
			Predicate: blankNodes[3],
			Object:    blankNodes[4],
		},
	}
	adjList = GetAdjacencyList(triples)
	// adjList must have exactly 3 keys (N0, N2, N4)
	if len(adjList) != 3 {
		t.Errorf("adjacency list for the given graph should've only one key. Found %v keys", len(adjList))
	}
	if nChildren := len(adjList[blankNodes[0]]); nChildren != 2 {
		t.Errorf("Node 0 should've exactly 2 children. Found %v children", nChildren)
	}
	// there aren't any neighbors for other nodes.
	for i := 1; i < len(blankNodes); i++ {
		if nChildren := len(adjList[blankNodes[i]]); nChildren > 0 {
			t.Errorf("N%v should have no neighbors. Found %v neighbors", i+1, nChildren)
		}
	}
}

func TestTopologicalSortTriples(t *testing.T) {
	nodes := getNBlankNodes(5)

	// TestCase 1: only a single triple in the list
	// The graph is as follows:
	//        (N1)
	// (N0) --------> (N2)
	triples := []*parser.Triple{
		{nodes[0], nodes[1], nodes[2]},
	}
	// but it doesn't exist in the keys of the map.
	sortedTriples, err := TopologicalSortTriples(triples)
	if err != nil {
		t.Errorf("unexpected parsing a single triple list. Error: %v", err)
	}
	if len(sortedTriples) != len(triples) {
		t.Errorf("sorted triples doesn't have a proper dimension. Expected %v triples, found %v triples", len(triples), len(sortedTriples))
	}

	// TestCase 2: another valid test-case where the input is a cyclic graph.
	/*
	           (N0)
	          / ^  \
	     (N1)/  |   \(N2)
	         \  |   /
	          v |  v
	           (N3)
	   Triples:
	       1. N0 -> N1 -> N3
	       2. N0 -> N2 -> N3
	       3. N3 -> N4 -> N0     // couldn't show N4 in the graph.
	*/
	nodes = getNBlankNodes(5)
	triples = []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[3]},
		{Subject: nodes[0], Predicate: nodes[2], Object: nodes[3]},
		{Subject: nodes[3], Predicate: nodes[4], Object: nodes[0]},
	}
	sortedTriples, err = TopologicalSortTriples(triples)

	// since we have a cycle here, we can expect two configurations.
	expectedTriples := []*parser.Triple{
		{Subject: nodes[3], Predicate: nodes[4], Object: nodes[0]},
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[3]},
		{Subject: nodes[0], Predicate: nodes[2], Object: nodes[3]},
	}
	anotherConfig := []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[3]},
		{Subject: nodes[0], Predicate: nodes[2], Object: nodes[3]},
		{Subject: nodes[3], Predicate: nodes[4], Object: nodes[0]},
	}
	if !reflect.DeepEqual(sortedTriples, expectedTriples) && !reflect.DeepEqual(sortedTriples, anotherConfig) {
		t.Errorf("sorted triples are not in correct order")
	}
}

func Test_topologicalSort(t *testing.T) {
	nodes := getNBlankNodes(5)

	// TestCase 1: invalid adjacency matrix with nodes which not in
	// the keys but exists in the children lists.
	// The graph is as follows:
	//        (N1)
	// (N0) --------> (N2)
	adjList := map[*parser.Node][]*parser.Node{
		nodes[0]: {nodes[2]},
	} // here, nodes[2] is child of nodes[0]
	// but it doesn't exist in the keys of the map.
	_, err := topologicalSort(adjList)
	if err == nil {
		t.Error("expected an error reporting \"extra nodes found\"")
	}

	// TestCase 2: Valid case
	adjList = map[*parser.Node][]*parser.Node{
		nodes[0]: {nodes[2]},
		nodes[2]: {},
	}
	sortedNodes, err := topologicalSort(adjList)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOutput := []*parser.Node{nodes[2], nodes[0]}
	if !reflect.DeepEqual(sortedNodes, expectedOutput) {
		t.Errorf("nodes are not sorted correctly. \nExpected: %v, \nFound: %v", sortedNodes, expectedOutput)
	}
}

func Test_topologicalSortHelper(t *testing.T) {
	// declaring satellite field required by the function
	var lastIndex int
	var adjList map[*parser.Node][]*parser.Node
	var visited map[*parser.Node]bool
	var resultList []*parser.Node

	reinitializeSatellites := func(nNodes int) {
		// nNodes := number of nodes.
		lastIndex = 0
		visited = make(map[*parser.Node]bool, nNodes)
		resultList = make([]*parser.Node, nNodes)
	}

	/*
	   Graph that we will be using for all the testcases in this function:
	    It is a simple three staged input with a single source and sink pair.

	                            (N1)
	                   (N0) ------------> (N2)
	                    |                  |
	                    |                  |
	                (N3)|                  |(N6)
	                    |                  |
	                    v                  v
	                   (N4) ------------> (N7)
	                           (N5)
	       Triples that exists in the above graph:
	       1. N0 -> N1 -> N2
	       2. N2 -> N6 -> N7
	       3. N0 -> N3 -> N4
	       4. N3 -> N5 -> N7
	*/
	numberNodes := 8
	nodes := getNBlankNodes(numberNodes)
	triples := []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[2]},
		{Subject: nodes[2], Predicate: nodes[6], Object: nodes[7]},
		{Subject: nodes[0], Predicate: nodes[3], Object: nodes[4]},
		{Subject: nodes[3], Predicate: nodes[5], Object: nodes[7]},
	}
	adjList = GetAdjacencyList(triples)

	// TestCase 1: trying to traverse on a node which doesn't exist in the graph.
	// function should raise an error.
	reinitializeSatellites(numberNodes)
	inexistentNode := parser.Node{NodeType: parser.BLANK, ID: "sample node"}
	err := topologicalSortHelper(&inexistentNode, &lastIndex, adjList, &visited, &resultList)
	if err == nil {
		t.Errorf("inexistent node should've raised an error")
	}

	// TestCase 2: traversing node with no children
	reinitializeSatellites(numberNodes)
	err = topologicalSortHelper(nodes[7], &lastIndex, adjList, &visited, &resultList)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// if a node without a child is traversed, the number of nodes added to the
	// resultList should be exactly 1. That is the node itself.
	if lastIndex != 1 {
		t.Errorf("Expected exactly 1 node in the result list. Found %v nodes", lastIndex)
	}

	// TestCase 3: traversing a node with 1 child
	// if we are traversing node N2, N7 is the only child.
	reinitializeSatellites(numberNodes)
	_ = topologicalSortHelper(nodes[2], &lastIndex, adjList, &visited, &resultList)
	if lastIndex != 2 {
		t.Errorf("Expected exactly 1 node in the result list. Found %v nodes", lastIndex)
	}
	// since N7 is the child of N2,
	// resultList must  have N7 before N2.
	if resultList[0] != nodes[7] || resultList[1] != nodes[2] {
		t.Error("resultList is not set properly")
	}

	// TestCase 4: Final Case when all the nodes will be traversed.
	reinitializeSatellites(numberNodes)
	_ = topologicalSortHelper(nodes[0], &lastIndex, adjList, &visited, &resultList)
	if lastIndex != 4 {
		t.Errorf("after parsing all nodes, resultList must've 4 nodes. Found %v nodes", lastIndex)
	}
	// after traversing all the nodes, (N0) should be the last node
	// to be added to the list and (N7) should be the first node.
	if resultList[0] != nodes[7] || resultList[lastIndex-1] != nodes[0] {
		t.Error("order of resultList if not correct")
	}
}

func TestDisjointSet(t *testing.T) {

	// TestCase 1: two distinct sets
	/* Modelling the following graph:
	            (N5)            (N6)
		(N0) ---------> (N1) ---------> (N2)

	           (N7)
	    (N3) --------> (N4)
	*/
	// Clear enough, there are two different sets in the depicted graph.
	nodes := getNBlankNodes(8)
	triples := []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[5], Object: nodes[1]},
		{Subject: nodes[1], Predicate: nodes[6], Object: nodes[2]},
		{Subject: nodes[3], Predicate: nodes[7], Object: nodes[4]},
	}
	parent := DisjointSet(triples)
	nSets := 0
	for subject := range parent {
		if parent[subject] == nil {
			nSets++
		}
	}
	if nSets != 2 {
		t.Errorf("Expected Graph to have exactly %d disjoint sets. Found %d sets", 2, nSets)
	}

	// TestCase 2: All Independent Triples:
	/*
	           (N8)                      (N9)
	   (N0) -----------> (N1)    (N2) -----------> (N3)

	           (N10)                     (N11)
	   (N4) -----------> (N5)    (N6) -----------> (N7)
	*/
	nodes = getNBlankNodes(12)
	triples = []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[8], Object: nodes[1]},
		{Subject: nodes[2], Predicate: nodes[9], Object: nodes[3]},
		{Subject: nodes[4], Predicate: nodes[10], Object: nodes[5]},
		{Subject: nodes[6], Predicate: nodes[11], Object: nodes[7]},
	}
	parent = DisjointSet(triples)
	nSets = 0
	for node := range parent {
		if parent[node] == nil {
			nSets++
		}
	}
	if nSets != len(triples) {
		t.Errorf("Mismatch in the number of sets. expected %v disjoint sets, found %v sets", len(triples), nSets)
	}
}

func Test_any(t *testing.T) {
	// any function checks if the target is in the given list of strings

	listStrings := []string{"c", "b", "a"}

	// case when the target string is present in the strings
	targetString := "a"
	if !any(targetString, listStrings) {
		// targetString wasn't found even though it exits in the list
		t.Errorf("couldn't find %v in %v. even though it exits in the list", targetString, listStrings)
	}

	// case when the target is not present in the strings
	targetString = "z"
	if any(targetString, listStrings) {
		// targetString was found even though it doesn't exist in the list.
		t.Errorf("found %v in the list: %v. even though it doesn't exist in the list", targetString, listStrings)
	}
}

func Test_invertSchemaDefinition(t *testing.T) {
	// no special variations to test for.
	spdxString := "http://spdx.org/rdf/terms"
	spdxURI, _ := uri.NewURIRef(parser.RDFNS)
	schemaDefinition := map[string]uri.URIRef{
		"spdx": spdxURI,
	}

	inv := invertSchemaDefinition(schemaDefinition)
	expected := map[string]string{spdxString: "spdx"}
	if reflect.DeepEqual(inv, expected) {
		t.Errorf("expected %v, found %v", expected, inv)
	}
}

func Test_getRDFNSAbbreviation(t *testing.T) {
	invSchemaDefinition := make(map[string]string)

	// TestCase 1: Default prefix of the rdf namespace should be "rdf"
	abbrev := getRDFNSAbbreviation(invSchemaDefinition)
	if abbrev != "rdf" {
		t.Errorf("by default, the rdf namespace should be abbreviated by rdf and not %s", abbrev)
	}

	// TestCase 2: Case when the rdf namespace exists in the invSchemaDefinition.
	expectedAbbrev := "rdfNS"
	invSchemaDefinition[parser.RDFNS] = expectedAbbrev
	abbrev = getRDFNSAbbreviation(invSchemaDefinition)
	if abbrev != expectedAbbrev {
		t.Errorf("expected %s abbreviation, found %v", expectedAbbrev, abbrev)
	}
}

func Test_getRestTriples(t *testing.T) {
	nodes := getNBlankNodes(7)

	// TestCase 1: Base Case where all the triples have predicate
	//             of rdf attributes
	triples := []*parser.Triple{
		{
			Subject:   nodes[0],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "type"},
			Object:    nodes[1]},
		{
			Subject:   nodes[2],
			Predicate: &parser.Node{NodeType: parser.IRI, ID: parser.RDFNS + "nodeID"},
			Object:    nodes[3],
		},
	}
	restTriples := getRestTriples(triples)
	if n := len(restTriples); n != 0 {
		t.Errorf("expected empty output. got %d nodes", n)
	}

	// TestCase 2: Case where the output has at least one node which is not a
	//             rdf attribute
	triples = append(triples, &parser.Triple{
		Subject:   nodes[4],
		Predicate: nodes[5],
		Object:    nodes[6],
	})
	restTriples = getRestTriples(triples)
	if n := len(restTriples); n != 1 {
		t.Errorf("expected only one extra triple. found %d triples", n)
	}
}

func Test_shortenURI(t *testing.T) {
	invSchema := getInvSchema()

	// TestCase 1: uri doesn't have two fragments
	//             Must raise an error
	uriref := "http://spdx.org/rdf/terms#"
	_, err := shortenURI(uriref, invSchema)
	if err == nil {
		t.Errorf("didn't raise any error for url with no fragment")
	}

	// TestCase 2: uri with inexistent baseURI
	//             Must raise an error
	uriref = "https://www.googlge.com/terms#website"
	_, err = shortenURI(uriref, invSchema)
	if err == nil {
		t.Errorf("expected an error stating baseURI doesn't exist in the given schema")
	}

	// TestCase 3: valid uri
	uriref = "http://spdx.org/rdf/terms#Snippet"
	expectedOP := "spdx:Snippet"
	shortURI, err := shortenURI(uriref, invSchema)
	if err != nil {
		t.Errorf("unexpected error converting a valid URI")
	}
	if shortURI != expectedOP {
		t.Errorf("expected output: %v, found: %v", expectedOP, shortURI)
	}
}

func Test_getRootNodes(t *testing.T) {
	nodes := getNBlankNodes(10)

	// TestCase 1: self loop
	triples := []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[0]},
	}
	roots := GetRootNodes(triples)
	if len(roots) != 0 {
		t.Errorf("expected no roots as output. found %v", roots)
	}

	// TestCase 2: Disjoint Triples with distinct root elements
	/*
	   Graph:  3 disjoint triplets

	                (N1)
	        (N0) ---------> (N2)

	               (N4)
	        (N3) ---------> (N5)

	               (N7)
	        (N6) ---------> (N8)
	*/
	triples = []*parser.Triple{
		{Subject: nodes[0], Predicate: nodes[1], Object: nodes[2]},
		{Subject: nodes[3], Predicate: nodes[4], Object: nodes[5]},
		{Subject: nodes[6], Predicate: nodes[7], Object: nodes[8]},
	}
	roots = GetRootNodes(triples)
	if len(roots) != len(triples) {
		t.Errorf("expected %v root nodes, found %v nodes", len(roots), len(triples))
	}
}
