package main

import (
	"fmt"
	"io"

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
	ymod = getMyYangModule(node)
	debuglog("generateField(): Generating for field %s.%s", node.NName(), node.Kind())
	switch node.Kind() {
	case "container":
		c, ok := node.(*yang.Container)
		if !ok {
			errorlog("generateField(): %s.%s not a container", node.NName(), node.Kind())
		}
		tn := fullName(c)
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fieldname))
		fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(fieldname), genTN(ymod, tn), nsstr, fieldname)
	case "leaf":
		l, ok := node.(*yang.Leaf)
		if !ok {
			errorlog("generateField(): %s.%s not a leaf", node.NName(), node.Kind())
		}
		tn := getTypeName(ymod, l.Type)
		pre := getPrefix(tn)
		if getImportedModuleByPrefix(ymod, pre) == nil {
			errorlog("generateField(): Exiting from leaf field: pre=%s, leaf=%s.%s", pre, node.NName(), node.Kind())
			break
		}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(fieldname))
		if l.Type != nil && l.Type.Name != "empty" {
			fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(fieldname), tn, nsstr, fieldname)
		}
	case "leaf-list":
		l, ok := node.(*yang.LeafList)
		if !ok {
			errorlog("generateField(): %s.%s not a leaf list", node.NName(), node.Kind())
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
			errorlog("generateField(): %s.%s not a list", node.NName(), node.Kind())
		}
		tn := fullName(l)
		fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(fieldname), genTN(ymod, tn), nsstr, fieldname)
	case "uses":
		u, ok := node.(*yang.Uses)
		if !ok {
			errorlog("generateField(): %s.%s not a uses", node.NName(), node.Kind())
		}
		pre := getPrefix(u.Name)
		if getImportedModuleByPrefix(ymod, pre) == nil {
			break
		}
		fmt.Fprintf(w, "\t%s\n", genTN(ymod, fieldname))
	default:
		errorlog("generateField(): unsupported field %s.%s", node.NName(), node.Kind())
	}
}

// This function goes through the list of entries that are contained within elements
// such as grouping, container, lists, etc. and generates the needed type definitions
func generateType(w io.Writer, ymod *yang.Module, node yang.Node, keepXmlID bool) {
	debuglog("generateTypes(): Generating type for %s", node.NName())
	switch node.Kind() {
	case "container":
		genTypeForContainer(w, ymod, node, keepXmlID)
	case "list":
		genTypeForList(w, ymod, node)
	case "leaf":
		genTypeForLeaf(w, ymod, node)
	case "leaf-list":
		genTypeForLeafList(w, ymod, node)
	}
}

