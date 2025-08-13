package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

// This function generates a single entry of field of a structure that may be generated
// from a compound structure such as a grouping, container, list, etc.
func generateField(w io.Writer, ymod *yang.Module, node yang.Node, prev yang.Node, addNs bool) {
	var nsstr string
	var fullname string
	if addNs {
		mod := getMyModule(ymod)
		nsstr = mod.namespace + " "
	}
	nodeName := node.NName()
	ymod = getMyYangModule(prev)
	debuglog("generateField(): Generating for field %s.%s", node.NName(), node.Kind())
	if node.ParentNode().Kind() == "augment" {
		fullname = fullName(prev) + "_" + node.NName()
	} else {
		fullname = fullName(node)
	}
	switch node.Kind() {
	case "container":
		//c, ok := node.(*yang.Container)
		//if !ok {
		//	errorlog("generateField(): %s.%s not a container", node.NName(), node.Kind())
		//}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(nodeName))
		fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(nodeName), genTN(ymod, fullname), nsstr, nodeName)
	case "notification":
		//notif, ok := node.(*yang.Notification)
		//if !ok {
		//	errorlog("generateField(): %s.%s not a notification", node.NName(), node.Kind())
		//}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(nodeName))
		fmt.Fprintf(w, "\t%s %s_cont `xml:\"%s%s\"`\n", genFN(nodeName), genTN(ymod, fullname), nsstr, nodeName)
	case "choice":
		//choice, ok := node.(*yang.Choice)
		//if !ok {
		//	errorlog("generateField(): %s.%s not a notification", node.NName(), node.Kind())
		//}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(nodeName))
		fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(nodeName), genTN(ymod, fullname), nsstr, nodeName)
	case "case":
		//case1, ok := node.(*yang.Case)
		//if !ok {
		//	errorlog("generateField(): %s.%s not a notification", node.NName(), node.Kind())
		//}
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(nodeName))
		fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(nodeName), genTN(ymod, fullname), nsstr, nodeName)
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
		fmt.Fprintf(w, "\t%s_Prsnt bool `xml:\",presfield\"`\n", genFN(nodeName))
		if l.Type != nil && l.Type.Name != "empty" {
			fmt.Fprintf(w, "\t%s %s `xml:\"%s%s\"`\n", genFN(nodeName), tn, nsstr, nodeName)
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
		fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(nodeName), tn, nsstr, nodeName)
	case "list":
		//l, ok := node.(*yang.List)
		//if !ok {
		//	errorlog("generateField(): %s.%s not a list", node.NName(), node.Kind())
		//}
		fmt.Fprintf(w, "\t%s []%s `xml:\"%s%s\"`\n", genFN(nodeName), genTN(ymod, fullname), nsstr, nodeName)
	case "uses":
		u, ok := node.(*yang.Uses)
		if !ok {
			errorlog("generateField(): %s.%s not a uses", node.NName(), node.Kind())
		}
		pre := getPrefix(u.Name)
		if getImportedModuleByPrefix(ymod, pre) == nil {
			break
		}
		fmt.Fprintf(w, "\t%s\n", genTN(ymod, nodeName))
	default:
		errorlog("generateField(): unsupported field %s.%s", nodeName,  node.Kind())
	}
}

// This function goes through the list of entries that are contained within elements
// such as grouping, container, lists, etc. and generates the needed type definitions
func generateType(w io.Writer, ymod *yang.Module, node yang.Node, prev yang.Node, keepXmlID bool) {
	debuglog("generateTypes(): Generating type for %s", node.NName())
	switch node.Kind() {
	case "container":
		genTypeForContainer(w, ymod, node, prev, keepXmlID)
	case "list":
		genTypeForList(w, ymod, node, prev)
	case "leaf":
		genTypeForLeaf(w, ymod, node, prev)
	case "leaf-list":
		genTypeForLeafList(w, ymod, node, prev)
	case "choice":
		genTypeForChoice(w, ymod, node, prev, keepXmlID)
	case "case":
		genTypeForCase(w, ymod, node, prev, keepXmlID)
	default:
		errorlog("generateType(): %s.%s is not yet supported", node.NName(), node.Kind())
	}
}

