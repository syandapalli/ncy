package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addChoiceComment(w io.Writer, c *yang.Container) {
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

func genTypeForChoice(w io.Writer, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	/*
	var name string
	var addNs bool = false
	c, ok := n.(*yang.Choice)
	if !ok {
		panic("Not a Choice")
	}

	addContainerComment(w, c)
	if c.ParentNode().Kind() != "augment" {
		name = fullName(c)
	} else {
		name = c.NName()
	}
	fmt.Fprintf(w, "type %s_cont struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, c.Name)
	}
	for _, c1 := range c.Container {
		generateField(w, ymod, c1, addNs)
	}
	for _, l1 := range c.Leaf {
		generateField(w, ymod, l1, addNs)
	}
	for _, g1 := range c.Grouping {
		generateField(w, ymod, g1, addNs)
	}
	for _, l1 := range c.List {
		generateField(w, ymod, l1, addNs)
	}
	for _, u1 := range c.Uses {
		generateField(w, ymod, u1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(c)
	generateContainerRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, c1 := range c.Container {
		if c1.ParentNode() == c {
			generateType(w, ymod, c1, false)
		}
	}
	for _, l1 := range c.Leaf {
		if l1.ParentNode() == c {
			generateType(w, ymod, l1, false)
		}
	}
	for _, l1 := range c.List {
		if l1.ParentNode() == c {
			generateType(w, ymod, l1, false)
		}
	}
	*/
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

