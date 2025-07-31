package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addAugmentToContainer(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentsToContainer(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	c, ok := n.(*yang.Container)
	if !ok {
		errorlog("addAugmentToContainer(): %s.%s is not a container", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		c.Container = append(c.Container, c1)
	}
	for _, l1 := range a.Leaf {
		c.Leaf = append(c.Leaf, l1)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToContainer(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			c.Container = append(c.Container, c1)
		}
		for _, l1 := range g.List {
			c.List = append(c.List, l1)
		}
		for _, l1 := range g.Leaf {
			c.Leaf = append(c.Leaf, l1)
		}
	}
}

func addAugmentToList(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentToList(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	l, ok := n.(*yang.List)
	if !ok {
		errorlog("addAugmentToList(): %s.%s is not a list", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		l.Container = append(l.Container, c1)
	}
	for _, l1 := range a.Leaf {
		l.Leaf = append(l.Leaf, l1)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToList(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			l.Container = append(l.Container, c1)
		}
		for _, l1 := range g.List {
			l.List = append(l.List, l1)
		}
		for _, l1 := range g.Leaf {
			l.Leaf = append(l.Leaf, l1)
		}
	}
}

func (mod *Module) preprocessAugment(aug *yang.Augment) {
	debuglog("preprocessAugment(): name=%s in module %s", aug.Name, mod.name)
	// Let's locate the position of the augment within the other module
	needleaf := false
	node := traverse(aug.Name, aug, needleaf)
	if node != nil {
		debuglog("preprocessAUgment(): found %s.%s for augment %s", node.NName(), node.Kind(), aug.NName())
		switch node.Kind() {
		case "container":
			addAugmentToContainer(aug, node)
		case "list":
			addAugmentToList(aug, node)
		default:
			errorlog("preprocessAugment(): addition to %s.%s not supported", node.NName(), node.Kind())
		}
	} else {
		errorlog("ERROR: Augment %s of module %s couldn't be located", aug.NName(), mod.name)
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

func processAugments(w io.Writer, submod *SubModule, ymod *yang.Module, n yang.Node) {
	a, ok := n.(*yang.Augment)
	if !ok {
		errorlog("processAugments(): %s.%s is not an Augment", n.NName(), n.Kind())
		return
	}

	for _, c := range a.Container {
		debuglog("processAugments(): generating for %s.%s in %s", c.NName(), c.Kind(), a.NName())
		/*
		genTypeForContainer(w, ymod, yang.Node(c), false)
		addAugmentComment(w, a)
		fmt.Fprintf(w, "type %s struct {\n", genAN(a.FullName()))
		generateAugments(w, ymod, a)
		fmt.Fprintf(w, "}\n")
		*/
	}
	for _, l := range a.Leaf {
		debuglog("processAugments(): generating for %s.%s in %s", l.NName(), l.Kind(), a.NName())
	}
	for _, l := range a.List {
		debuglog("processAugments(): generating for %s.%s in %s", l.NName(), l.Kind(), a.NName())
	}
}
