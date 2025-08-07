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
	for _, n1 := range l.Notification {
		generateField(w, m, n1, addNs)
	}
	for _, c1 := range l.Choice {
		generateField(w, m, c1, addNs)
	}
	for _, u1 := range l.Uses {
		generateField(w, m, u1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// The code below generates the type definitions needed
	// for the constituents inside a list
	for _, cont := range l.Container {
		if cont.ParentNode() == l {
			generateType(w, m, cont, false)
		}
	}
	for _, leaf := range l.Leaf {
		if leaf.ParentNode() == l {
			generateType(w, m, leaf, false)
		}
	}
	for _, list := range l.List {
		if list.ParentNode() == l {
			generateType(w, m, list, false)
		}
	}
	for _, notif := range l.Notification {
		if notif.ParentNode() == l {
			generateType(w, m, notif, false)
		}
	}
	for _, choice := range l.Choice {
		if choice.ParentNode() == l {
			generateType(w, m, choice, false)
		}
	}
}

// Look for a node that belongs to the list with a specific name. Iterate through
// the fields, match the field name to the passed name and return if it matches.
// It is different for any field that has 'uses' syntax. For such, we iterate through
// the fields of the uses structure and identify the match.
func getNodeFromList(l *yang.List, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromList(): looking for %s in %s", fname, l.NName())
	name := getName(fname)
	for _, c1 := range l.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range l.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range l.LeafList {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range l.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, c1 := range l.Choice {
		if c1.NName() == name {
			return c1
		}
	}
	for _, u1 := range l.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	return nil
}

// This function attempts to locate a uses node within the list recursively
// till it finds a uses node whic uses the same string as passed above.
// TODO: prefix handling must be properly handled
func getMatchingUsesNodeFromList(l *yang.List, name string) yang.Node {
	for _, u1 := range l.Uses {
		uname := getName(u1.NName())
		iname := getName(name)
		if uname == iname {
			return l
		}
	}
	//for _, g1 := range l.Grouping {
	//	if n := getMatchingUsesNodeFromGrouping(g1, name); n != nil {
	//		return n
	//	}
	//}
	for _, c1 := range l.Container {
		if n := getMatchingUsesNodeFromContainer(c1, name); n != nil {
			return n
		}
	}
	for _, l1 := range l.List {
		if n := getMatchingUsesNodeFromList(l1, name); n != nil {
			return n
		}
	}
	for _, c1 := range l.Choice {
		if n := getMatchingUsesNodeFromChoice(c1, name); n != nil {
			return n
		}
	}
	for _, n1 := range l.Notification {
		if n := getMatchingUsesNodeFromNotification(n1, name); n != nil {
			return n
		}
	}
	return nil
}
