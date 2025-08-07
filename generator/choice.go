package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addChoiceComment(w io.Writer, c *yang.Choice) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(c.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	s = indentString(c.Description.Name)
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

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

func genTypeForChoice(w io.Writer, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	var name string
	var addNs bool = false
	choice, ok := n.(*yang.Choice)
	if !ok {
		errorlog("genTypeForChoice(): %s.%s is not a Choice", n.NName(), n.Kind())
		return
	}

	addChoiceComment(w, choice)
	if choice.ParentNode().Kind() != "augment" {
		name = fullName(choice)
	} else {
		name = choice.NName()
	}
	fmt.Fprintf(w, "type %s struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, choice.NName())
	}
	for _, cont := range choice.Container {
		generateField(w, ymod, cont, addNs)
	}
	for _, leaf := range choice.Leaf {
		generateField(w, ymod, leaf, addNs)
	}
	for _, list := range choice.List {
		generateField(w, ymod, list, addNs)
	}
	for _, case1 := range choice.Case {
		generateField(w, ymod, case1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(choice)
	generateChoiceRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, cont := range choice.Container {
		if cont.ParentNode() == choice {
			generateType(w, ymod, cont, false)
		}
	}
	for _, leaf := range choice.Leaf {
		if leaf.ParentNode() == choice {
			generateType(w, ymod, leaf, false)
		}
	}
	for _, list := range choice.List {
		if list.ParentNode() == choice {
			generateType(w, ymod, list, false)
		}
	}
	for _, case1 := range choice.Case {
		if case1.ParentNode() == choice {
			generateType(w, ymod, case1, false)
		}
	}
}

func genTypeForCase(w io.Writer, ymod *yang.Module, n yang.Node, keepXmlID bool) {
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
		name = case1.NName()
	}
	fmt.Fprintf(w, "type %s struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, case1.NName())
	}
	for _, cont := range case1.Container {
		generateField(w, ymod, cont, addNs)
	}
	for _, leaf := range case1.Leaf {
		generateField(w, ymod, leaf, addNs)
	}
	for _, list := range case1.List {
		generateField(w, ymod, list, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(case1)
	generateChoiceRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, cont := range case1.Container {
		if cont.ParentNode() == case1 {
			generateType(w, ymod, cont, false)
		}
	}
	for _, leaf := range case1.Leaf {
		if leaf.ParentNode() == case1 {
			generateType(w, ymod, leaf, false)
		}
	}
	for _, list := range case1.List {
		if list.ParentNode() == case1 {
			generateType(w, ymod, list, false)
		}
	}
}

func generateChoiceRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, name string) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, name))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}

func getNodeFromChoice(c *yang.Choice, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromChoice(): looking for %s in %s", fname, c.NName())
	name := getName(fname)
	for _, c1 := range c.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range c.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range c.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, c1 := range c.Case {
		if c1.NName() == name {
			return c1
		}
	}
	errorlog("getNodeFromChoice(): failed to find %s in %s.%s", name, c.NName(), c.Kind())
	return nil
}

func getNodeFromCase(c *yang.Case, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromCase(): looking for %s in %s", fname, c.NName())
	name := getName(fname)
	for _, c1 := range c.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range c.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range c.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, u1 := range c.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	errorlog("getNodeFromCase(): failed to find %s in %s.%s", name, c.NName(), c.Kind())
	return nil
}

// This function attempts to locate a uses node within the container recursively
// till it finds a uses node whic uses the same string as passed above.
// TODO: prefix handling must be properly handled
func getMatchingUsesNodeFromChoice(c *yang.Choice, name string) yang.Node {
	debuglog("getMatchingUsesNodeFromChoice(): looking for %s in %s.%s", name, c.NName(), c.Kind())
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
	for _, c1 := range c.Case {
		if n := getMatchingUsesNodeFromCase(c1, name); n != nil {
			return n
		}
	}
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

