package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func (mod *Module) preprocessAugment(aug *yang.Augment) {
	//fmt.Println("Augment Name", aug.Name, "in module", mod.name)
	// Let's locate the position of the augment within the other module
	needleaf := false
	node := traverse(aug.Name, aug, needleaf)
	if node != nil {
		/*
		c, ok := node.(*yang.Container)
		if ok {
			// Add the augment to the container so that when code is
			// generated for the container, the augments are used in
			// field generation. TODO - commented to pass the compilation
			c.AddAugment(aug)
		} else {
			panic("preprocessAugment() - Node located isn't a container: " + nodeString(node))
		}
		*/
		fmt.Println("Hit this case that was commented out")
	} else {
		fmt.Println("ERROR: Augment couldn't be located")
	}
}

func (mod *Module) preprocessAugments() {
	for _, sm := range mod.submodules {
		ymod := sm.module
		for _, aug := range ymod.Augment {
			mod.preprocessAugment(aug)
		}
	}
}

// This function checks if a grouping is part of any augments that
// are declared within this module or submodule
func groupingInAugment(ymod *yang.Module, g *yang.Grouping) bool {
	for _, a := range ymod.Augment {
		for _, u := range a.Uses {
			if u.Name == g.Name {
				return true
			}
		}
	}
	return false
}

func addAugmentComment(w io.Writer, a *yang.Augment) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString(a.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	s = indentString(a.Description.Name)
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

func processAugments(w io.Writer, mod *Module, ymod *yang.Module, n yang.Node) {
	a, ok := n.(*yang.Augment)
	if !ok {
		panic("Not an Augment")
	}

	for _, c := range a.Container {
		genTypeForContainer(w, ymod, yang.Node(c), false)
	}
	/*
		addAugmentComment(w, a)
		fmt.Fprintf(w, "type %s struct {\n", genAN(a.FullName()))
		generateAugments(w, ymod, a)
		fmt.Fprintf(w, "}\n")
	*/
}
