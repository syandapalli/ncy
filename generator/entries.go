package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

// This function generates a single entry of field of a structure that may be generated
// from a compound structure such as a grouping, container, list, etc.
func generateField(w io.Writer, ymod *yang.Module, node yang.Node, addNs bool) {
	var nsstr string
	if addNs {
		mod := getMyModule(ymod)
		nsstr = mod.namespace + " "
	}
	fieldname := node.NName()
	debuglog("Generating for field %s", fieldname)
	switch node.Kind() {
	case "container":
		c, ok := node.(*yang.Container)
		if !ok {
			panic("Not a container")
		}
		tn := getFullName(c)
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fieldname))
		fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fieldname), genTN(ymod, tn), nsstr, fieldname)
	case "leaf":
		l, ok := node.(*yang.Leaf)
		if !ok {
			panic("Not a leaf")
		}
		tn := getTypeName(ymod, l.Type)
		pre := getPrefix(getType(ymod, l.Type))
		if getImportedModuleByPrefix(ymod, pre) == nil {
			errorlog("Exiting from leaf field: pre=%s", pre)
			break
		}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fieldname))
		if l.Type != nil && l.Type.Name != "empty" {
			fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fieldname), tn, nsstr, fieldname)
		}
	case "leaf-list":
		l, ok := node.(*yang.LeafList)
		if !ok {
			panic("Not a LeafList")
		}
		tn := getTypeName(ymod, l.Type)
		pre := getPrefix(getType(ymod, l.Type))
		if getImportedModuleByPrefix(ymod, pre) == nil {
			break
		}
		fmt.Fprintf(w, "// Generated from here pre = %s, tn = %s \n", pre, l.Type.Name)
		fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fieldname), tn, nsstr, fieldname)
	case "list":
		l, ok := node.(*yang.List)
		if !ok {
			panic("Not a Leaf")
		}
		tn := getFullName(l)
		fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fieldname), genTN(ymod, tn), nsstr, fieldname)
	case "uses":
		u, ok := node.(*yang.Uses)
		if !ok {
			panic("Not a Uses")
		}
		pre := getPrefix(u.Name)
		if getImportedModuleByPrefix(ymod, pre) == nil {
			break
		}
		fmt.Fprintf(w, "\t%s\n", genTN(ymod, fieldname))
	default:
		errorlog("in generation of field for %s", node.Kind())
	}
}

// This function goes through the list of entries that are contained within elements
// such as grouping, container, lists, etc. and generates the needed type definitions
func generateTypes(w io.Writer, m *yang.Module, n yang.Node, keepXmlID bool) {
	debuglog("Generating type for %s", n.NName())
	switch n.Kind() {
	case "container":
		genTypeForContainer(w, m, n, keepXmlID)
	case "list":
		genTypeForList(w, m, n)
	case "leaf":
		genTypeForLeaf(w, m, n)
	case "leaf-list":
		genTypeForLeafList(w, m, n)
	}
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
		errorlog("Augment case not supported yet: %s", nodeContextStr(aug))
	}
}
