package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

// This function generates fields based on entries in the grouping, container
// or list as each of them can have multiple container, list and leaf elements
// within them. This function doesn't generate type defitions needed.
func generateFields(w io.Writer, ymod *yang.Module, a *yang.Container, addNs bool) {
	/*
	var nsstr string
	var es []yang.Node
	// TODO: Needs to be fixed. Changed only for compilation
	// es := a.GetEntries()
	es = a.Leaf
	if addNs {
		mod := getMyModule(ymod)
		nsstr = mod.namespace + " "
	}
	for _, e := range es {
		fn := e.NName()
		switch e.Kind() {
		case "container":
			c, ok := e.(*yang.Container)
			if !ok {
				panic("Not a container")
			}
			tn := c.NName()
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fn), genTN(ymod, tn), nsstr, fn)
		case "leaf":
			l, ok := e.(*yang.Leaf)
			if !ok {
				panic("Not a leaf")
			}
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			if l.Type != nil && l.Type.Name != "empty" {
				fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
			}
		case "leaf-list":
			l, ok := e.(*yang.LeafList)
			if !ok {
				panic("Not a LeafList")
			}
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "// Generated from here pre = %s, tn = %s \n", pre, l.Type.Name)
			fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
		case "list":
			l, ok := e.(*yang.List)
			if !ok {
				panic("Not a Leaf")
			}
			tn := l.NName()
			fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), genTN(ymod, tn), nsstr, fn)
		case "uses":
			u, ok := e.(*yang.Uses)
			if !ok {
				panic("Not a Uses")
			}
			pre := getPrefix(u.Name)
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "\t%s\n", genTN(ymod, fn))
		}
	}
	for _, e := range es {
		fn := e.NName()
		switch e.Kind() {
		case "container":
			c, ok := e.(*yang.Container)
			if !ok {
				panic("Not a container")
			}
			tn := c.NName()
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fn), genTN(ymod, tn), nsstr, fn)
		case "leaf":
			l, ok := e.(*yang.Leaf)
			if !ok {
				panic("Not a leaf")
			}
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			if l.Type != nil && l.Type.Name != "empty" {
				fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
			}
		case "leaf-list":
			l, ok := e.(*yang.LeafList)
			if !ok {
				panic("Not a LeafList")
			}
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "// Generated from here pre = %s, tn = %s\n", pre, l.Type.Name)
			fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
		case "list":
			l, ok := e.(*yang.List)
			if !ok {
				panic("Not a Leaf")
			}
			tn := l.NName()
			fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), genTN(ymod, tn), nsstr, fn)
		case "uses":
			u, ok := e.(*yang.Uses)
			if !ok {
				panic("Not a Uses")
			}
			pre := getPrefix(u.Name)
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "\t%s\n", genTN(ymod, fn))
		}
	}
	var augmentsAdded = map[string]bool{}
	for _, e := range a.GetAugments() {
		augMod := getMyModule(e)
		nsstr := augMod.namespace + " "
		augYangMod := getMyYangModule(e)
		aug, ok := e.(*yang.Augment)
		if !ok {
			panic("The node is not an Augment: " + nodeString(e))
		}

		switch {
		case len(aug.Uses) > 0:
			// Lets figure out the type name to be used when including the uses
			// in the code generated. If the name includes no ":", alternatively
			// no scope, the scope name (prefix) is derived from the module where
			// the augment is defined
			for _, u := range aug.Uses {
				tname := u.Name
				if !strings.Contains(tname, ":") {
					tname = augMod.prefix + ":" + tname
				}
				if _, ok := augmentsAdded[tname]; ok {
					continue
				} else {
					augmentsAdded[tname] = true
				}
				fmt.Fprintf(w, "\t%s\n", genFN(tname))
			}
		case len(aug.Leaf) > 0:
			for _, l := range aug.Leaf {
				fn := l.NName()
				tname := fn
				if !strings.Contains(tname, ":") {
					tname = augMod.prefix + ":" + tname
				}
				if _, ok := augmentsAdded[tname]; ok {
					continue
				} else {
					augmentsAdded[tname] = true
				}
				tn := getTypeName(augYangMod, l.Type)
				pre := getPrefix(getType(augYangMod, l.Type))
				if getImportedModuleByPrefix(augYangMod, pre) == nil {
					break
				}
				fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
				fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
			}
		case len(aug.LeafList) > 0:
			for _, l := range aug.LeafList {
				fn := l.NName()
				tname := fn
				if !strings.Contains(tname, ":") {
					tname = augMod.prefix + ":" + tname
				}
				if _, ok := augmentsAdded[tname]; ok {
					continue
				} else {
					augmentsAdded[tname] = true
				}
				tn := getTypeName(augYangMod, l.Type)
				pre := getPrefix(getType(augYangMod, l.Type))
				if getImportedModuleByPrefix(augYangMod, pre) == nil {
					break
				}
				fmt.Fprintf(w, "// Generated from here pre = %s, tn = %s\n", pre, l.Type.Name)
				fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
			}
		case len(aug.Container) > 0:
			for _, c := range aug.Container {
				fn := c.NName()
				tname := fn
				if !strings.Contains(tname, ":") {
					tname = augMod.prefix + ":" + tname
				}
				if _, ok := augmentsAdded[tname]; ok {
					continue
				} else {
					augmentsAdded[tname] = true
				}
				fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
				fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fn), genTN(augYangMod, fn), nsstr, fn)
			}

		default:
			fmt.Println("ERROR: Augment case not supported yet:", nodeStringFull(aug))
		}
	}
	*/
}

// This function goes through the list of entries that are contained within elements
// such as grouping, container, lists, etc. and generates the needed type definitions
func generateTypes(w io.Writer, m *yang.Module, c *yang.Container, keepXmlID bool) {
	/*
	entries := c.GetEntries()
	for _, e := range entries {
		switch e.Kind() {
		case "container":
			processContainer(w, m, e, keepXmlID)
		case "list":
			processList(w, m, e)
		case "leaf":
			processLeaf(w, m, e)
		case "leaf-list":
			processLeaflist(w, m, e)
		}
	}
	*/
}

func generateAugments(w io.Writer, ymod *yang.Module, aug *yang.Augment) {
	mod := getMyModule(ymod)
	nsstr := mod.namespace + " "
	switch {
	case len(aug.Uses) > 0:
		for _, u := range aug.Uses {
			tname := u.Name
			if !strings.Contains(tname, ":") {
				tname = ymod.Prefix.Name + ":" + tname
			}
			fmt.Fprintf(w, "\t%s\n", genFN(tname))
		}
	case len(aug.Leaf) > 0:
		for _, l := range aug.Leaf {
			fn := l.NName()
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
		}
	case len(aug.LeafList) > 0:
		for _, l := range aug.LeafList {
			fn := l.NName()
			tn := getTypeName(ymod, l.Type)
			pre := getPrefix(getType(ymod, l.Type))
			if getImportedModuleByPrefix(ymod, pre) == nil {
				break
			}
			fmt.Fprintf(w, "/* Generated from here pre = %s, tn = %s */\n", pre, l.Type.Name)
			fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fn), tn, nsstr, fn)
		}
	case len(aug.Container) > 0:
		for _, c := range aug.Container {
			fn := c.NName()
			fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fn))
			fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fn), genTN(ymod, fn), nsstr, fn)
		}
	default:
		fmt.Println("ERROR: Augment case not supported yet:", nodeStringFull(aug))
	}
}
