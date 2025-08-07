package main

import (
	//"fmt"
	//"io"

	"github.com/openconfig/goyang/pkg/yang"
)

/*
func addContainerComment(w io.Writer, c *yang.Container) {
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

func genTypeForContainer(w io.Writer, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	var name string
	var addNs bool = false
	c, ok := n.(*yang.Container)
	if !ok {
		panic("Not a Container")
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
}

func generateContainerRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, name string) {
	fmt.Fprintf(w, "func (x %s_cont) RuntimeNs() string {\n", genTN(ymod, name))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}
*/

func getNodeFromNotification(n *yang.Notification, fname string, leaf bool) yang.Node {
	debuglog("getNodeFromContainer(): looking for %s in %s", fname, n.NName())
	name := getName(fname)
	for _, c1 := range n.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range n.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range n.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, c1 := range n.Choice {
		if c1.NName() == name {
			return c1
		}
	}
	for _, u1 := range n.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	return nil
}

// This function attempts to locate a uses node within the container recursively
// till it finds a uses node whic uses the same string as passed above.
// TODO: prefix handling must be properly handled
func getMatchingUsesNodeFromNotification(n *yang.Notification, name string) yang.Node {
	for _, u1 := range n.Uses {
		uname := getName(u1.NName())
		iname := getName(name)
		if uname == iname {
			return n
		}
	}
	for _, g1 := range n.Grouping {
		if n := getMatchingUsesNodeFromGrouping(g1, name); n != nil {
			return n
		}
	}
	for _, c1 := range n.Container {
		if n := getMatchingUsesNodeFromContainer(c1, name); n != nil {
			return n
		}
	}
	for _, l1 := range n.List {
		if n := getMatchingUsesNodeFromList(l1, name); n != nil {
			return n
		}
	}
	for _, c1 := range n.Choice {
		if n := getMatchingUsesNodeFromChoice(c1, name); n != nil {
			return n
		}
	}
	return nil
}

