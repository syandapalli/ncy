package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

// Case statements are part of choice statements.
// This function generates comments for case statement as a part of generation
func addCaseComment(w io.Writer, case1 *yang.Case) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(case1.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	if case1.Description != nil {
		s = indentString(case1.Description.Name)
		s = commentString(s)
		fmt.Fprint(w, s)
	}
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

// Each case statement is translated into a go type which is then
// used to create fields in the structure generated for choice statement
// The case may contain almost any other statement except for case statements
func genTypeForCase(w io.Writer, ymod *yang.Module, n yang.Node, prev yang.Node, keepXmlID bool) {
	var name string
	var addNs bool = false
	case1, ok := n.(*yang.Case)
	if !ok {
		errorlog("genTypeForCase(): %s.%s is not a case", n.NName(), n.Kind())
		return
	}

	addCaseComment(w, case1)
	if case1.ParentNode().Kind() != "augment" {
		name = fullName(case1)
	} else {
		name = fullName(prev) + "_" + case1.NName()
	}
	fmt.Fprintf(w, "type %s struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, case1.NName())
	}
	for _, cont := range case1.Container {
		generateField(w, ymod, cont, case1, addNs)
	}
	for _, leaf := range case1.Leaf {
		generateField(w, ymod, leaf, case1, addNs)
	}
	for _, list := range case1.List {
		generateField(w, ymod, list, case1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(case1)
	generateChoiceRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, cont := range case1.Container {
		if cont.ParentNode() == case1 {
			generateType(w, ymod, cont, case1, false)
		}
	}
	for _, leaf := range case1.Leaf {
		if leaf.ParentNode() == case1 {
			generateType(w, ymod, leaf, case1, false)
		}
	}
	for _, list := range case1.List {
		if list.ParentNode() == case1 {
			generateType(w, ymod, list, case1, false)
		}
	}
}

// Look for a node with the name passed as a parameter of the function
// The fname may be of form prefix:name. We first pick out the name from
// the fname and match only with name. TODO: In theory even the prefix
// must be matched but is ignored for now.
func getNodeFromCase(case1 *yang.Case, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromCase(): looking for %s in %s", fname, case1.NName())
	name := getName(fname)
	for _, c1 := range case1.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range case1.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range case1.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, u1 := range case1.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	errorlog("getNodeFromCase(): failed to find %s in %s.%s", name, case1.NName(), case1.Kind())
	return nil
}

func getMatchingUsesNodeFromCase(c *yang.Case, name string) yang.Node {
	debuglog("getMatchingUsesNodeFromCase(): looking for %s in %s.%s", name, c.NName(), c.Kind())
	for _, c1 := range c.Container {
		if n := getMatchingUsesNodeFromContainer(c1, name); n != nil {
			return n
		}
	}
	for _, l1 := range c.List {
		if n := getMatchingUsesNodeFromList(l1, name); n != nil {
			return n
		}
	}
	for _, c1 := range c.Choice {
		if n := getMatchingUsesNodeFromChoice(c1, name); n != nil {
			return n
		}
	}
	return nil
}

