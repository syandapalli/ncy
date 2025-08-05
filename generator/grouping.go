package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addGroupingComment(w io.Writer, g *yang.Grouping) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(g.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	s = indentString(g.Description.Name)
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

func printType(w io.Writer, t *yang.Type) {
	fmt.Fprintf(w, "Name: %s\n", t.Name)
	if len(t.Type) > 0 {
		printType(w, t.Type[0])
	}
}

func processGrouping(w io.Writer, submod *SubModule, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	// Check the precondtions before we dive deep in
	g, ok := n.(*yang.Grouping)
	if !ok {
		errorlog("processGrouping(): Not a Grouping:%s", n.NName())
	}

	// Add comment to describe the source of the generated code
	addGroupingComment(w, g)

	// If the grouping is part of any augment, we need to add namespace
	// for each field of the grouping. Check if the uses is included in
	// any augment
	addNs := groupingInAugment(ymod, g)

	// The code below generates code for the grouping
	debuglog("processGrouping(): Generating for group %s", g.NName())
	fmt.Fprintf(w, "type %s struct {\n", genTN(ymod, g.NName()))
	for _, l1 := range g.Leaf {
		generateField(w, ymod, l1, addNs)
	}
	for _, c1 := range g.Container {
		generateField(w, ymod, c1, addNs)
	}
	for _, g1 := range g.Grouping {
		generateField(w, ymod, g1, addNs)
	}
	for _, l1 := range g.List {
		generateField(w, ymod, l1, addNs)
	}
	for _, u1 := range g.Uses {
		generateField(w, ymod, u1, addNs)
	}
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	generateGroupingRuntimeNs(w, submod, ymod, g)

	// The code below triggers the code generation for the
	// constituents of the grouping
	for _, l1 := range g.Leaf {
		if l1.ParentNode() == g {
			generateType(w, ymod, l1, addNs)
		}
	}
	for _, c1 := range g.Container {
		if c1.ParentNode() == g {
			generateType(w, ymod, c1, addNs)
		}
	}
	for _, l1 := range g.List {
		if l1.ParentNode() == g {
			generateType(w, ymod, l1, addNs)
		}
	}
	fmt.Fprintf(w, "\n")

	//storeInGroupingMap(submod.prefix, n)
}

// Namespace is an important aspect of NC/XML. This function allows us to return
// namespace for the structures we generate in a granular fashion.
func generateGroupingRuntimeNs(w io.Writer, submod *SubModule, ymod *yang.Module, g *yang.Grouping) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, g.NName()))
	mod := getMyModule(ymod)
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}

// Get a node with a specific name from within the grouping. We iterate through
// the fields and return if the name matches. The treatment of 'uses' is different
// as the fields of grouping included using 'uses' is inserted as is the node where
// it is included. Thus, for 'uses', we need to iterate through the fields of the
// included grouping.
// TODO. The iteration for 'uses' is not implemented yet
func getNodeFromGrouping(n yang.Node, name string, leaf bool) yang.Node {
	debuglog("getNodeFromGrouping(): looking for %s in %s.%s", name, n.NName(), n.Kind())
	g, ok := n.(*yang.Grouping)
	if !ok {
		errorlog("a non grouping passed: %s", g.Kind())
		return nil
	}
	for _, c1 := range g.Container {
		if c1.NName() == name {
			return c1
		}
	}
	for _, l1 := range g.Leaf {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range g.List {
		if l1.NName() == name {
			return l1
		}
	}
	for _, l1 := range g.LeafList {
		if l1.NName() == name {
			return l1
		}
	}
	for _, u1 := range g.Uses {
		if node := getNodeFromUses(u1, name); node != nil {
			return node
		}
	}
	return nil
}

// This is used to locate a grouping based on where it is instantiated
// Typically a grouping is instantiated using "uses" construct of YANG
// The inclusion may be recursively anywhere inside the hierarchical nature
// of YANG structures. We search recursively till we locate the node
func getMatchingUsesNodeFromGrouping(g *yang.Grouping, name string) yang.Node {
	for _, u1 := range g.Uses {
		uname := getName(u1.NName())
		iname := getName(name)
		if uname == iname {
			return g
		}
	}
	for _, g1 := range g.Grouping {
		if n := getMatchingUsesNodeFromGrouping(g1, name); n != nil {
			return n
		}
	}
	for _, c1 := range g.Container {
		if n := getMatchingUsesNodeFromContainer(c1, name); n != nil {
			return n
		}
	}
	for _, l1 := range g.List {
		if n := getMatchingUsesNodeFromList(l1, name); n != nil {
			return n
		}
	}
	return nil
}


// One of the utility functions that help traversal across the YANG specification
func getGroupingByName(u *yang.Uses) *yang.Grouping {
	prefix := getPrefix(u.NName())
	gname := getName(u.NName())
	ymod := getMyYangModule(u)
	if prefix != "" {
		ymod = getImportedYangModuleByPrefix(ymod, prefix)
	}
	if ymod == nil {
		errorlog("getGroupingByName(): module not found for prefix=%s, mod=%s", prefix, ymod.NName())
		return nil
	}
	for _, g := range ymod.Grouping {
		if g.NName() == gname {
			return g
		}
	}
	errorlog("getGroupingByName():Unable to locate grouping %s in module %s", u.NName(), ymod.NName())
	return nil
}
