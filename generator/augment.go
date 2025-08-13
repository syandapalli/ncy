package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

// The following functions make sure the same element is not added twice.
// If the same element is added twice, the code generation fails.
func addContainer(c *yang.Container, cs []*yang.Container) []*yang.Container {
	for _, x := range cs {
		if x == c {
			return cs
		}
	}
	return append(cs, c)
}
func addList(l *yang.List, ls []*yang.List) []*yang.List {
	for _, x := range ls {
		if x == l {
			return ls
		}
	}
	return append(ls, l)
}
func addLeaf(l *yang.Leaf, ls []*yang.Leaf) []*yang.Leaf {
	for _, x := range ls {
		if x == l {
			return ls
		}
	}
	return append(ls, l)
}
func addChoice(c *yang.Choice, cs []*yang.Choice) []*yang.Choice {
	for _, x := range cs {
		if x == c {
			return cs
		}
	}
	return append(cs, c)
}
func addNotification(n *yang.Notification, ns []*yang.Notification) []*yang.Notification {
	for _, x := range ns {
		if x == n {
			return ns
		}
	}
	return append(ns, n)
}
func addCase(c *yang.Case, cs []*yang.Case) []*yang.Case {
	for _, x := range cs {
		if x == c {
			return cs
		}
	}
	return append(cs, c)
}

// Each augment is added to a some container type. Each augment contains multiple statements
// or elements that get added to whereever the augment points to. Thus, as a result the
// augment is added to a variety of container type elements


// Add all the elements of a container
func addAugmentToContainer(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentsToContainer(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	cont, ok := n.(*yang.Container)
	if !ok {
		errorlog("addAugmentToContainer(): %s.%s is not a container", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		cont.Container = addContainer(c1, cont.Container) 
	}
	for _, l1 := range a.Leaf {
		cont.Leaf = addLeaf(l1, cont.Leaf)
	}
	for _, c1 := range a.Choice {
		cont.Choice = addChoice(c1, cont.Choice)
	}
	for _, n1 := range a.Notification {
		cont.Notification = addNotification(n1, cont.Notification)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToContainer(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			cont.Container = addContainer(c1, cont.Container)
		}
		for _, l1 := range g.List {
			cont.List = addList(l1, cont.List)
		}
		for _, l1 := range g.Leaf {
			cont.Leaf = addLeaf(l1, cont.Leaf)
		}
	}
}

// Add elements of augment to the list
func addAugmentToList(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentToList(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	list, ok := n.(*yang.List)
	if !ok {
		errorlog("addAugmentToList(): %s.%s is not a list", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		list.Container = addContainer(c1, list.Container)
	}
	for _, l1 := range a.Leaf {
		list.Leaf = addLeaf(l1, list.Leaf)
	}
	for _, l1 := range a.List {
		list.List = addList(l1, list.List)
	}
	for _, c1 := range a.Choice {
		list.Choice = addChoice(c1, list.Choice)
	}
	for _, n1 := range a.Notification {
		list.Notification = addNotification(n1, list.Notification)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToList(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			list.Container = addContainer(c1, list.Container)
		}
		for _, l1 := range g.List {
			list.List = addList(l1, list.List)
		}
		for _, l1 := range g.Leaf {
			list.Leaf = addLeaf(l1, list.Leaf)
		}
	}
}

// add elements of augment to notification
func addAugmentToNotification(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentToNotification(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	notif, ok := n.(*yang.Notification)
	if !ok {
		errorlog("addAugmentToList(): %s.%s is not a list", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		notif.Container = addContainer(c1, notif.Container)
	}
	for _, l1 := range a.Leaf {
		notif.Leaf = addLeaf(l1, notif.Leaf)
	}
	for _, l1 := range a.List {
		notif.List = addList(l1, notif.List)
	}
	for _, c1 := range a.Choice {
		notif.Choice = addChoice(c1, notif.Choice)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToList(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			notif.Container = addContainer(c1, notif.Container)
		}
		for _, l1 := range g.List {
			notif.List = addList(l1, notif.List)
		}
		for _, l1 := range g.Leaf {
			notif.Leaf = addLeaf(l1, notif.Leaf)
		}
	}
}

// Add elements of augment to choice
func addAugmentToChoice(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentToChoice(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	choice, ok := n.(*yang.Choice)
	if !ok {
		errorlog("addAugmentToList(): %s.%s is not a choice", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		choice.Container = addContainer(c1, choice.Container)
	}
	for _, l1 := range a.Leaf {
		choice.Leaf = addLeaf(l1, choice.Leaf)
	}
	for _, l1 := range a.List {
		choice.List = addList(l1, choice.List)
	}
	for _, c1 := range a.Case {
		choice.Case = addCase(c1, choice.Case)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToList(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			choice.Container = addContainer(c1, choice.Container)
		}
		for _, l1 := range g.List {
			choice.List = addList(l1, choice.List)
		}
		for _, l1 := range g.Leaf {
			choice.Leaf = addLeaf(l1, choice.Leaf)
		}
	}
}

// Add elements of augment to case
func addAugmentToCase(a *yang.Augment, n yang.Node) {
	debuglog("addAugmentToChoice(): adding %s to %s.%s", a.NName(), n.NName(), n.Kind())
	case1, ok := n.(*yang.Case)
	if !ok {
		errorlog("addAugmentToList(): %s.%s is not a case", n.NName(), n.Kind())
	}
	for _, c1 := range a.Container {
		case1.Container = addContainer(c1, case1.Container)
	}
	for _, l1 := range a.Leaf {
		case1.Leaf = addLeaf(l1, case1.Leaf)
	}
	for _, l1 := range a.List {
		case1.List = addList(l1, case1.List)
	}
	for _, c1 := range a.Choice {
		case1.Choice = addChoice(c1, case1.Choice)
	}
	for _, u1 := range a.Uses {
		g := getGroupingByName(u1)
		if g == nil {
			errorlog("addAugmentToList(): couldn't locate grouping %s", u1.NName())
			continue
		}
		for _, c1 := range g.Container {
			case1.Container = addContainer(c1, case1.Container)
		}
		for _, l1 := range g.List {
			case1.List = addList(l1, case1.List)
		}
		for _, l1 := range g.Leaf {
			case1.Leaf = addLeaf(l1, case1.Leaf)
		}
	}
}

// preprocess augment aims to traverse through the path provided to
// identify the container type element where the contents of the 
// the augment are to be placed
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
		case "choice":
			addAugmentToChoice(aug, node)
		case "notification":
			addAugmentToNotification(aug, node)
		default:
			errorlog("preprocessAugment(): addition to %s.%s not supported", node.NName(), node.Kind())
		}
	} else {
		errorlog("ERROR: Augment %s of module %s couldn't be located", aug.NName(), mod.name)
	}
}

// Preprocess augments of a module which includes processing them from
// the submodules that are part of the module
func (mod *Module) preprocessAugments() {
	// the submodules must be processed in order for the traversal to work.
	// The includes list in the module has the order of the processing of the
	// submodules
	for _, inc := range mod.module.Include {
		sm, ok := mod.submodules[inc.NName()]
		if !ok {
			errorlog("preprocessAugments(): couldn't find submodule %s", inc.NName())
			continue
		}
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
/*
	a, ok := n.(*yang.Augment)
	if !ok {
		errorlog("processAugments(): %s.%s is not an Augment", n.NName(), n.Kind())
		return
	}

	for _, c := range a.Container {
		debuglog("processAugments(): generating for %s.%s in %s", c.NName(), c.Kind(), a.NName())
		addAugmentComment(w, a)
		genTypeForContainer(w, ymod, yang.Node(c), nil, false)
	}
	for _, l := range a.Leaf {
		debuglog("processAugments(): generating for %s.%s in %s", l.NName(), l.Kind(), a.NName())
		addAugmentComment(w, a)
		genTypeForLeaf(w, ymod, l, nil)
	}
	for _, l := range a.List {
		debuglog("processAugments(): generating for %s.%s in %s", l.NName(), l.Kind(), a.NName())
		addAugmentComment(w, a)
		genTypeForList(w, ymod, l, nil)
	}
*/
}

