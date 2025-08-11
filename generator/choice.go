package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

// Generate comment for the structure that will be generated for the
// choice. The comments include some information that was used in generation
// so that it is possible to debug too
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

// Generate the structure that represents the choice. We traverse through
// the fields of the choice and generate code for each element. Mostly, it 
// must be purely case statements. However, it is legal to have other types
// of statements instead of case statements.
func genTypeForChoice(w io.Writer, ymod *yang.Module, n yang.Node, prev yang.Node, keepXmlID bool) {
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
		name = fullName(prev) + "_" + choice.NName()
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
		generateType(w, ymod, cont, choice, false)
	}
	for _, leaf := range choice.Leaf {
		generateType(w, ymod, leaf, choice, false)
	}
	for _, list := range choice.List {
		generateType(w, ymod, list, choice, false)
	}
	for _, case1 := range choice.Case {
		generateType(w, ymod, case1, choice, false)
	}
}

// Generate runtime namespace for the structure. This is used by
// the encoder to see when the namesapce is changed and the transistion
// must be recorded in the encoding
func generateChoiceRuntimeNs(w io.Writer, mod *Module, ymod *yang.Module, name string) {
	fmt.Fprintf(w, "func (x %s) RuntimeNs() string {\n", genTN(ymod, name))
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mod.name))
	fmt.Fprintf(w, "}\n")
}

// This function is used within the traversal of path either as a part of augment
// or other concepts such as leafref, etc. This function looks to locate a field
// of the same name as passed in the parameters
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

