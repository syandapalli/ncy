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

func processGrouping(w io.Writer, mod *Module, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	// Check the precondtions before we dive deep in
	g, ok := n.(*yang.Grouping)
	if !ok {
		panic("Not Grouping.")
	}

	// Add comment to describe the source of the generated code
	addGroupingComment(w, g)

	// If the grouping is part of any augment, we need to add namespace
	// for each field of the grouping. Check if the uses is included in
	// any augment
	addNs := groupingInAugment(ymod, g)

	// The code below generates code for the grouping
	fmt.Fprintf(w, "type %s struct {\n", genTN(ymod, g.NName()))
	generateFields(w, ymod, g, addNs)
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	generateRuntimeNs(w, mod, ymod, g)

	// The code below triggers the code generation for the
	// constituents of the grouping
	generateTypes(w, ymod, g, keepXmlID)
	fmt.Fprintf(w, "\n")

	storeInGroupingMap(mod.prefix, n)
}

func generateRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, g *yang.Grouping) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, g.NName()))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}
