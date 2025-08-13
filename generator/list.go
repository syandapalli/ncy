package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func genTypeForList(w io.Writer, m *yang.Module, n yang.Node, prev yang.Node) {
	var addNs bool = false
	var ln string
	// Complete some sanity checks before going ahead
	list, ok := n.(*yang.List)
	if !ok {
		errorlog("genTypeForList(): %s.%s is not a List", n.NName(), n.Kind())
		return
	}

	// We are all good. Let's generate the first type
	// that represents the list which has fields which
	// are also generated within
	if list.ParentNode().Kind() != "augment" {
		ln = fullName(list)
	} else {
		ln = fullName(prev) + "_" + list.NName()
	}

	// Now start generating the code for the list
	fmt.Fprintf(w, "type %s struct {\n", genTN(m, ln))
	for _, l1 := range list.Leaf {
		generateField(w, m, l1, list, addNs)
	}
	for _, c1 := range list.Container {
		generateField(w, m, c1, list, addNs)
	}
	for _, g1 := range list.Grouping {
		generateField(w, m, g1, list, addNs)
	}
	for _, l1 := range list.List {
		generateField(w, m, l1, list, addNs)
	}
	for _, n1 := range list.Notification {
		generateField(w, m, n1, list, addNs)
	}
	for _, c1 := range list.Choice {
		generateField(w, m, c1, list, addNs)
	}
	for _, u1 := range list.Uses {
		generateField(w, m, u1, list, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// The code below generates the type definitions needed
	// for the constituents inside a list
	for _, cont := range list.Container {
		generateType(w, m, cont, list, false)
	}
	for _, leaf := range list.Leaf {
		generateType(w, m, leaf, list, false)
	}
	for _, list1 := range list.List {
		generateType(w, m, list1, list, false)
	}
	for _, notif := range list.Notification {
		generateType(w, m, notif, list, false)
	}
	for _, choice := range list.Choice {
		generateType(w, m, choice, list, false)
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
