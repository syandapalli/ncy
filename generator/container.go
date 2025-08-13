package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

// Generate comments for the structure that is generated for the container
// The comments include information that may be used during debugging too and
// used by developers to understand the source of the generated code
func addContainerComment(w io.Writer, c *yang.Container) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(c.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	if c.Description != nil {
		s = indentString(c.Description.Name)
		s = commentString(s)
		fmt.Fprint(w, s)
	}
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

// Generate structure for the container which essentially is a set of fields
// which are elements of the container.
func genTypeForContainer(w io.Writer, ymod *yang.Module, n yang.Node, prev yang.Node, keepXmlID bool) {
	var name string
	var addNs bool = false
	cont, ok := n.(*yang.Container)
	if !ok {
		errorlog("getTypeForContainer(): %s.%s is not a Container", n.NName(), n.Kind())
		return
	}

	// Find out some useufl information in generation of the name of the
	// structure that is generated for the container
	addContainerComment(w, cont)
	if cont.ParentNode().Kind() != "augment" {
		name = fullName(cont)
	} else {
		name = fullName(prev) + "_" + cont.NName()
	}

	// Now we start generating code for the container
	fmt.Fprintf(w, "type %s_cont struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, cont.Name)
	}
	for _, c1 := range cont.Container {
		generateField(w, ymod, c1, cont, addNs)
	}
	for _, l1 := range cont.Leaf {
		generateField(w, ymod, l1, cont, addNs)
	}
	for _, g1 := range cont.Grouping {
		generateField(w, ymod, g1, cont, addNs)
	}
	for _, l1 := range cont.List {
		generateField(w, ymod, l1, cont, addNs)
	}
	for _, n1 := range cont.Notification {
		generateField(w, ymod, n1, cont, addNs)
	}
	for _, c1 := range cont.Choice {
		generateField(w, ymod, c1, cont, addNs)
	}
	for _, u1 := range cont.Uses {
		generateField(w, ymod, u1, cont, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(cont)
	generateContainerRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, cont1 := range cont.Container {
		generateType(w, ymod, cont1, cont, false)
	}
	for _, leaf := range cont.Leaf {
		generateType(w, ymod, leaf, cont, false)
	}
	for _, list := range cont.List {
		generateType(w, ymod, list, cont, false)
	}
	for _, notif := range cont.Notification {
		generateType(w, ymod, notif, cont, false)
	}
	for _, choice := range cont.Choice {
		generateType(w, ymod, choice, cont, false)
	}
}

func generateContainerRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, name string) {
	fmt.Fprintf(w, "func (x %s_cont) RuntimeNs() string {\n", genTN(ymod, name))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}

func getNodeFromContainer(c *yang.Container, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromContainer(): looking for %s in %s", fname, c.NName())
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
	for _, c1 := range c.Choice {
		if c1.NName() == name {
			return c1
		}
	}
	for _, u1 := range c.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	return nil
}

// This function attempts to locate a uses node within the container recursively
// till it finds a uses node whic uses the same string as passed above.
// TODO: prefix handling must be properly handled
func getMatchingUsesNodeFromContainer(c *yang.Container, name string) yang.Node {
	for _, u1 := range c.Uses {
		uname := getName(u1.NName())
		iname := getName(name)
		if uname == iname {
			return c
		}
	}
	for _, g1 := range c.Grouping {
		if n := getMatchingUsesNodeFromGrouping(g1, name); n != nil {
			return n
		}
	}
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
	for _, n1 := range c.Notification {
		if n := getMatchingUsesNodeFromNotification(n1, name); n != nil {
			return n
		}
	}
	return nil
}

