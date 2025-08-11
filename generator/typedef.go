package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addComment(w io.Writer, typedef *yang.Typedef) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString("typedef: " + typedef.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	if typedef.Description != nil {
		s = indentString(typedef.Description.Name)
		s = commentString(s)
		fmt.Fprint(w, s)
	}
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

func processTypedef(w io.Writer, submod *SubModule, ymod *yang.Module, n yang.Node) {
	typedef, ok := n.(*yang.Typedef)
	if !ok {
		errorlog("processTypedef(): %s.%s not a typedef", n.NName(), n.Kind())
		return
	}
	addComment(w, typedef)
	processType(w, ymod, typedef.Type)
	generateTypedefRuntimeNs(w, submod, ymod, typedef)
	fmt.Fprintf(w, "\n")
}

func generateTypedefRuntimeNs(w io.Writer, submod *SubModule, ymod *yang.Module, typedef *yang.Typedef) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, typedef.NName()))
	mod := getMyModule(ymod)
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}
