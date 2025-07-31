package main

import (
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func getLeafTypeName(m *yang.Module, l *yang.Leaf) string {
	switch l.Type.Name {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		if l.Type.Range == nil {
			return l.Type.Name
		} else {
			return genTN(m, l.NName())
		}
	case "string":
		if l.Type.Length == nil {
			return "string"
		} else {
			return genTN(m, l.NName())
		}
	case "leafref":
		ref := getLeafref(l.Type.Path.Name, m, l)
		if ref == nil {
			errorlog("Couldn't locate the leafref")
		}
		return getLeafTypeName(m, ref)
	case "boolean":
		return "bool"
	default:
		return genTN(m, l.Type.Name)
	}
}

func genTypeForLeaf(w io.Writer, m *yang.Module, n yang.Node) {
	l, ok := n.(*yang.Leaf)
	if !ok {
		errorlog("Not a Leaf")
	}

	// Need to generate type for a leaf only if it creates a
	// data type using enumeration. The other types of must
	// be handled in future.
	processType(w, m, l.Type)
}

func genTypeForLeafList(w io.Writer, m *yang.Module, n yang.Node) {
	l, ok := n.(*yang.LeafList)
	if !ok {
		errorlog("Not a LeafList")
	}

	// Need to generate type for a leaf only if it creates a
	// data type using enumeration. The other types of must
	// be handled in future.
	processType(w, m, l.Type)
}

// Find a leaf for a leafref which may even be recursive. Traverse
// till you find a node that isn't leafref
func getLeafref(path string, m *yang.Module, n yang.Node) *yang.Leaf {
	// Traverse the tree to fetch the node pointed to by the leafref.
	debuglog("getLeafref(): locating path=%s for %s.%s", path, n.NName(), n.Kind())
	node := traverse(path, n, true)
	if node == nil {
		errorlog("getLeafref(): Failed to find leaf with reference path = %s, leaf = %s", path, n.NName())
		return nil
	}
	l, ok := node.(*yang.Leaf)
	if !ok {
		errorlog("getLeafref(): Not a leaf %s for path %s", n.NName(), path)
		return nil
	}

	// Here we handle the case of a leafref pointing to another leafref and
	// so on. This lower portion addresses recursive traversal to locate the
	// final leaf of interest
	if l.Type.Name == "leafref" {
		debuglog("getLeafref(): Found leafref with path=%s for path=%s", l.Type.Path.Name, path)
		return getLeafref(l.Type.Path.Name, m, l)
	}
	return l
}
