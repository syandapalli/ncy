package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func genTypeForList(w io.Writer, m *yang.Module, n yang.Node) {
	var addNs bool = false
	// Complete some sanity checks before going ahead
	l, ok := n.(*yang.List)
	if !ok {
		panic("Not a List")
	}

	// We are all good. Let's generate the first type
	// that represents the list which has fields which
	// are also generated within
	ln := fullName(l)
	fmt.Fprintf(w, "type %s struct {\n", genTN(m, ln))
	for _, l1 := range l.Leaf {
		generateField(w, m, l1, addNs)
	}
	for _, c1 := range l.Container {
		generateField(w, m, c1, addNs)
	}
	for _, g1 := range l.Grouping {
		generateField(w, m, g1, addNs)
	}
	for _, l1 := range l.List {
		generateField(w, m, l1, addNs)
	}
	for _, u1 := range l.Uses {
		generateField(w, m, u1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// The code below generates the type definitions needed
	// for the constituents inside a list
	for _, c1 := range l.Container {
		generateTypes(w, m, c1, false)
	}
	for _, l1 := range l.Leaf {
		generateTypes(w, m, l1, false)
	}
	for _, l1 := range l.List {
		generateTypes(w, m, l1, false)
	}
}

// Look for a node that belongs to the list with a specific name. Iterate through
// the fields, match the field name to the passed name and return if it matches.
// It is different for any field that has 'uses' syntax. For such, we iterate through
// the fields of the uses structure and identify the match.
func getNodeFromList(mod *Module, l *yang.List, name string, leaf bool) yang.Node {
	for _, c1 := range l.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, c1 := range l.Leaf {
		if c1.NName() == name {
			return c1
		}
	}
	for _, c1 := range l.LeafList {
		if c1.NName() == name {
			return c1
		}
	}
	for _, u1 := range l.Uses {
		if node := getNodeFromUses(mod, u1, name); node != nil {
			return node
		}
	}
	return nil
}

func getNodeWithUsesFromList(l *yang.List, name string) yang.Node {
	for _, u1 := range l.Uses {
		if u1.NName() == name {
			return l
		}
	}
	for _, g1 := range l.Grouping {
		if n := getNodeWithUsesFromGrouping(g1, name); n != nil {
			return n
		}
	}
	for _, c1 := range l.Container {
		if n := getNodeWithUsesFromContainer(c1, name); n != nil {
			return n
		}
	}
	for _, l1 := range l.List {
		if n := getNodeWithUsesFromList(l1, name); n != nil {
			return n
		}
	}
	return nil
}
