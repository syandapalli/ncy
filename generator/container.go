package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

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

func processContainer(w io.Writer, ymod *yang.Module, n yang.Node, keepXmlID bool) {
	var addNs bool = false
	c, ok := n.(*yang.Container)
	if !ok {
		panic("Not a Container")
	}

	addContainerComment(w, c)
	name := c.NName()
	if strings.Contains(c.NName(), "/") {
		// This container is inside augment. Use NName instead.
		name = c.NName()
	}
	fmt.Fprintf(w, "type %s_cont struct {\n", genTN(ymod, name))
	if keepXmlID {
		mod := getMyModule(ymod)
		fmt.Fprintf(w, "\tXMLName nc.XmlId `xml:\"%s %s\"`\n", mod.namespace, c.Name)
	}
	generateFields(w, ymod, c, addNs)
	fmt.Fprintf(w, "}\n")

	// Generate runtime namespace function
	mod := getMyModule(c)
	generateContainerRuntimeNs(w, mod, ymod, name)

	// The code below triggers the code generation for the
	// constituents of the grouping
	generateTypes(w, ymod, c, false)
}

func generateContainerRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, name string) {
	fmt.Fprintf(w, "func (x %s_cont) RuntimeNs() string {\n", genTN(ymod, name))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}
