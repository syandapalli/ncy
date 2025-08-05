package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addComment(w io.Writer, t *yang.Typedef) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(t.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	s = indentString(t.Description.Name)
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

func processTypedef(w io.Writer, submod *SubModule, ymod *yang.Module, n yang.Node) {
	t, ok := n.(*yang.Typedef)
	if !ok {
		panic("Not a typedef")
	}
	addComment(w, t)
	processType(w, ymod, t.Type)
	generateTypedefRuntimeNs(w, submod, ymod, t)
	fmt.Fprintf(w, "\n")
}

func generateTypedefRuntimeNs(w io.Writer, submod *SubModule, ymod *yang.Module, t *yang.Typedef) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, t.NName()))
	mod := getMyModule(ymod)
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}
