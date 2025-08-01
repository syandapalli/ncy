package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	//"strconv"

	"github.com/openconfig/goyang/pkg/yang"

	"regexp"
	"strings"
)


// ****************************************************************
// Utilities for managing the debugging. It is possible to turn the
// debug on/off but the error logs are always emitted
var debugenabled bool = false
//var debugenabled bool = true
func debuglog(format string, args ...interface{}) {
	if debugenabled {
		fmt.Printf("DEBUG: " + format + "\n", args...)
	}
}
func errorlog(format string, args ...interface{}) {
	fmt.Printf("ERROR: " + format + "\n", args...)
}

// A set of maps that are used to store and retrieve effieciently
// the modules and submodules.
var prefixModulesMap = map[string][]*yang.Module{}
var prefixModuleMap = map[string]*yang.Module{}
var groupingMap = map[string]yang.Node{}


// ***********************************************************
// Set of print functions that are useful when debugging
func printIndent(indent int) {
	for indent > 0 {
		fmt.Print("\t")
		indent = indent -1
	}
}

func printNode(n yang.Node, indent int) {
	printIndent(indent)
	fmt.Println(n.Kind(), n.NName())
	switch n.Kind() {
	case "container":
		c := n.(*yang.Container)
		for _, l := range c.Leaf {
			printNode(l, indent + 1)
		}
		for _, g := range c.Grouping {
			printNode(g, indent + 1)
		}
		for _, u := range c.Uses {
			printNode(u, indent + 1)
		}
		for _, l := range c.List {
			printNode(l, indent + 1)
		}
	case "grouping":
		g := n.(*yang.Grouping)
		for _, l := range g.Leaf {
			printNode(l, indent + 1)
		}
		for _, g := range g.Grouping {
			printNode(g, indent + 1)
		}
		for _, c := range g.Container {
			printNode(c, indent + 1)
		}
		for _, u := range g.Uses {
			printNode(u, indent + 1)
		}
		for _, l := range g.List {
			printNode(l, indent + 1)
		}
	case "list":
		l := n.(*yang.List)
		for _, lf := range l.Leaf {
			printNode(lf, indent + 1)
		}
		for _, g := range l.Grouping {
			printNode(g, indent + 1)
		}
		for _, c := range l.Container {
			printNode(c, indent + 1)
		}
		for _, u := range l.Uses {
			printNode(u, indent + 1)
		}
		for _, ls := range l.List {
			printNode(ls, indent + 1)
		}
	}
}
func printYangModule(m *yang.Module, indent int) {
	printIndent(indent)
	fmt.Println("Yang Module:", m.NName())
	for _, l := range m.Leaf {
		printNode(l, indent + 1)
	}
	for _, c := range m.Container {
		printNode(c, indent + 1)
	}
	for _, g := range m.Grouping {
		printNode(g, indent + 1)
	}
	for _, a := range m.Augment {
		printNode(a, indent + 1)
	}
	for _, u := range m.Uses {
		printNode(u, indent + 1)
	}
}

// Function that emits the name of the node in a short form
// to include in debug statements rather efficiently
func nodeString(node yang.Node) string {
	return node.Kind() + ":" + node.NName()
}

// Function to fully describe the node
func nodeContextStr(node yang.Node) string {
	var mname string
	ymod := getMyYangModule(node)
	if ymod.BelongsTo != nil {
		mname = ymod.BelongsTo.Name
	} else {
		mname = ymod.Name
	}

	return node.Kind() + ":" + node.NName() + "@ module:" + mname
}

// This function adds indentation to the text. Essentially adds a tab
// at the beginning of each new line
func indentString(c string) string {
	s := strings.ReplaceAll(c, "\n", "\n    ")
	s = "    " + s
	return s
}

// This utility converts a large text to comments by
// adding '//' in front of each line
func commentString(s string) string {
	return "//" + strings.ReplaceAll(s, "\n", "\n// ") + "\n"
}

// Get the prefix of a yang module. The location of information
// may depend on whether it is module or a submodule
func getYangPrefix(m *yang.Module) string {
	if m.BelongsTo != nil {
		return m.BelongsTo.Prefix.Name
	} else {
		return m.Prefix.Name
	}
}

// Get the prefix that is used by the main yang module. To get
// this we must get the module name first and look up the prefix
// from the main module
func getModulePrefix(m *yang.Module) string {
	var mname string
	if m.BelongsTo != nil {
		mname = m.BelongsTo.Name
	} else {
		mname = m.Name
	}
	mod, ok := modulesByName[mname]
	if ok {
		return mod.prefix
	}
	return ""
}

// 
func fullName(n yang.Node) string {
	fn := n.NName()
	for n.ParentNode() != nil && n.ParentNode().Kind() != "module" {
		n = n.ParentNode()
		fn = n.NName() + "_" + fn
	}
	return fn
}

// Generates a suitable type name to be used. If the type is
// a built-in, it translates the type to a golang equivalent.
// For other types, it ensures that right prefix is placed to
// make the type name unique. The prefix is derived from the
// module the definition belongs to
func genTN(m *yang.Module, s string) string {
	// This is a constructed type and needs to be handled
	if !strings.Contains(s, ":") {
		s = getModulePrefix(m) + ":" + s
	} else {
		pre := getPrefix(s)
		name := getName(s)
		pre = translatePrefix(m, pre)
		s = pre + ":" + name
	}
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, ".", "_")
	x := []byte(strings.ReplaceAll(s, "-", "_"))
	if x[0] > 96 {
		x[0] = x[0] - 32
	}
	return string(x)
}

// Generates field name based on the string passed. It
// modifies all '-'s to '_'s and changes the first character
// to upper case if it isn't already. It makes some
// assumptions that the text passed is from a proper yang
// file so the first character is always a lowercase or an
// uppercase alphabet
func genFN(s string) string {
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, ".", "_")
	x := []byte(strings.ReplaceAll(s, "-", "_"))
	if x[0] > 96 {
		x[0] = x[0] - 32
	}
	return string(x)
}

/*
// generate a suitable name for augment
func genAN(s string) string {
	// This is a constructed type and needs to be handled
	parts := strings.Split(s, "/")

	names := []string{}
	for _, p := range parts {
		y := strings.Split(p, ":")
		names = append(names, y[len(y)-1])
	}
	x := names[0]
	for _, n := range names[1:] {
		x = x + "_" + n
	}
	x = strings.ReplaceAll(x, "-", "_")
	x = strings.ReplaceAll(x, "__", "_")
	b := []byte(x)
	if b[0] > 96 {
		b[0] = b[0] - 32
	}

	return string(b) + "_augment"
}
*/

// Essentially, splits the string on ":" and provides the first
// part as prefix. If the string has no ":", prefix is returned
// as empty string
func getPrefix(s string) string {
	if !strings.Contains(s, ":") {
		return ""
	}
	parts := strings.Split(s, ":")
	return parts[0]
}

// Gets the non-prefix part which is essentially name of something
// The function assumes a good string is passed and doesn't check
// if the string passed has more than one ':'
func getName(s string) string {
	if !strings.Contains(s, ":") {
		return s
	}
	parts := strings.Split(s, ":")
	return parts[1]
}

// Returns the name of the module from the prefix. The search for
// the name must include evaluation of if the prefix is of the same
// module. If not, look through the imports.
func getModuleNameFromPrefix(ym *yang.Module, pre string) string {
	myprefix := getYangPrefix(ym)
	if pre == "" || pre == myprefix {
		if ym.BelongsTo != nil {
			return ym.BelongsTo.Name
		}
		return ym.Name
	}
	for _, imp := range ym.Import {
		if pre == imp.Prefix.Name {
			return imp.Name
		}
	}
	return ""
}

// This function gets yang.Module from any node by traversing up the
// tree. The final node should be of type *yang.Module.
func getMyYangModule(node yang.Node) *yang.Module {
	n := node
	for n.ParentNode() != nil {
		n = n.ParentNode()
	}
	ymod, ok := n.(*yang.Module)
	if !ok {
		errorlog("getMyYangModule(): module not found for %s.%s", node.NName(), node.Kind())
		return nil
	}
	return ymod
}

// This function is used to locate module from yang.Module. yang.Module
// is provided by the goyang module and "Module" is from this tool
func getMyModule(node yang.Node) *Module {
	var mname string
	ymod := getMyYangModule(node)
	if ymod == nil {
		panic("getMyModule() - yang module is nil")
	}
	if ymod.BelongsTo != nil {
		mname = ymod.BelongsTo.Name
	} else {
		mname = ymod.Name
	}
	mod, ok := modulesByName[mname]
	if !ok {
		errorlog("getMyModule() - couldn't locate module for %s.%s", node.NName(), node.Kind())
		return nil
	}
	return mod
}

// Get the imports module based on the prefix value passed. The prefix
// is used to find out the name of the module/submodule being referred
// to in the module passed as first argument. The actual module/submodule
// is found using the name.
// Note: getSubModule() fetches both main and submodules.
func getImportedModuleByPrefix(ymod *yang.Module, pre string) *Module {
	myprefix := getYangPrefix(ymod)
	if pre == "" || pre == myprefix {
		return getMyModule(ymod)
	}
	for _, i := range ymod.Import {
		if pre == i.Prefix.Name {
			// Now we found the name of the module. Locate the
			// module by name and return it
			mod, ok := modulesByName[i.Name]
			if ok {
				return mod
			}
		}
	}
	return nil
}

func getImportedYangModuleByPrefix(ymod *yang.Module, pre string) *yang.Module {
	myprefix := getYangPrefix(ymod)
	if pre == "" || pre == myprefix {
		return ymod
	}
	for _, i := range ymod.Import {
		if pre == i.Prefix.Name {
			// We found the yang module's name. We now
			// need to lcoate it
			if mod, ok := modulesByName[i.Name]; ok {
				for _, sm := range mod.submodules {
					if sm.module.NName() == i.Name {
						return sm.module
					}
				}
			}
		}
	}
	return nil
}

// Translation of prefix is to locate the imported module and
// pick up the prefix used in the module. The expectation is
// that the prefixes of all modules that form the yang specification
// are unique and there is no collision
func translatePrefix(m *yang.Module, pre string) string {
	target := getImportedModuleByPrefix(m, pre)
	if target != nil {
		return target.prefix
	}
	return ""
}

// Finds a node in the module by name. This traverses the first
// level of the module alone. This function is used by function
// getFromUsesNode() which is tasked to find a node that may be
// a grouping (a type included in other module) or a node within
// another uses in the top level of the module
/*
func getNodeFromModule(mod *Module, name string, needleaf bool) yang.Node {
	for _, sm := range mod.submodules {
		// TODO Incomplete
		for _, e := range sm.module.Leaf {
			if e.NName() == name {
				if !needleaf {
					return e
				} else if e.Kind() == "leaf" {
					return e
				}
			}
		}
	}
	return nil
}
*/

// Traverse a "Uses" for locating the node. The traversal could be
// recursive. Such a traversal can switch the module and thus we
// need to return both module and node to further the traversal thereon
func getFromUsesNode(n yang.Node, name string, needleaf bool) (*Module, yang.Node) {
	/*
	debuglog("getFromUsesNode() - Looking for %s in %s.%s", name, n.NName(), n.Kind()) 
	u, ok := n.(*yang.Uses)
	if !ok {
		panic("Not a Uses node")
	}

	// The uses may point to a different module which is determined
	// based on the prefix qualification in the "uses" name.
	prefix := getPrefix(u.Name)
	gname := getName(u.Name)
	m := getMyYangModule(u)
	modname := getModuleNameFromPrefix(m, prefix)
	mod, ok := modulesByName[modname]
	if !ok {
		errorlog("getFromUsesNode(): Module not found for prefix %s in module %s", prefix, m.Name)
		return nil, nil
	}
	node := getNodeFromModule(mod, gname, false)
	if node == nil {
		return mod, nil
	}
	if c, ok := node.(*yang.Container); ok {
		return mod, getNodeFromContainer(c, name, needleaf)
	}
	*/
	errorlog("getFromUsesNode(): returning nil as the it isn't a container")
	return nil, nil
}

// This utility locates the next node by name and returns it
// If the current node includes a uses, it must traverse the
// entries within the type pointed by uses. Thus, the traversal
// can be recursive.
func getNextNodeByName(node yang.Node, name string, leaf bool) yang.Node {
	mod := getMyModule(node)
	debuglog("getNextNodeByName():Finding %s curr=%s.%s", name, node.NName(), node.Kind())
	switch node.Kind() {
	case "grouping":
		return getNodeFromGrouping(node, name, false)
	case "container":
		return getNodeFromContainer(node.(*yang.Container), name, false)
	case "list":
		return getNodeFromList(node.(*yang.List), name, false)
	case "module":
		return getNodeFromMod(mod, name)
	}
	errorlog("getNextNodeByName(): %s.%s isn't a container type node", node.NName(), node.Kind())
	return nil
}

func getNodeFromMod(mod *Module, name string) yang.Node {
	debuglog("getNodeFromMod(): Getting \"%s\" from module %s", name, mod.name)
	for _, sm := range mod.submodules {
		ymod := sm.module
		for _, c1 := range ymod.Container  {
			if c1.NName() == name {
				return c1
			}
		}
		for _, u1 := range ymod.Uses {
			if n := getNodeFromUses(u1, name); n != nil {
				return n
			}
		}
	}
	errorlog("getNodeFromMod(): Failed to get %s from module %s", name, mod.name)
	return nil
}

// Locate grouping by its name from a module and return it
func getGroupingFromMod(mod *Module, name string) yang.Node {
	debuglog("getGroupingFromMod(): Getting \"%s\" from module %s", name, mod.name)
	for _, sm := range mod.submodules {
		ymod := sm.module
		for _, g1 := range ymod.Grouping  {
			if g1.NName() == name {
				return g1
			}
		}
	}
	errorlog("getGroupingFromMod(): Failed to get %s from module %s", name, mod.name)
	return nil
}

// Get the name of the uses and the corresponding grouping from
// the module if module is mentioned in the name. If not, the
// current module should be used to locate the grouping
func getNodeFromUses(u *yang.Uses, name string) yang.Node {
	var mod *Module
	debuglog("getNodeFromUses(): Getting %s from uses %s", name, u.NName())
	prefix := getPrefix(u.NName())
	uname := getName(u.NName())
	ymod := getMyYangModule(u)
	if prefix != "" {
		mod = getImportedModuleByPrefix(ymod, prefix)
		if mod == nil {
			errorlog("getNodeFromUses(): didn't locate module for prefix=%s module=%s", prefix, ymod.NName())
			return nil
		}
	} else {
		mod = getMyModule(u)
		if mod == nil {
			errorlog("getNodeFromUses(): didn't locate my module for %s.%s", u.NName(), u.Kind())
			return nil
		}
	}

	// Now locate the grouping and traverse it for the node with the 
	// 'name' string passed in the parameters
	// TODO: Move to *yang.Module instead of *Module
	if grouping := getGroupingFromMod(mod, uname); grouping != nil {
		return getNodeFromGrouping(grouping, name, false)
	}
	return nil
}

// Locate the node from the entire set of modules from the uses.
func getMatchingUsesNode(name string) yang.Node {
	for _, mod := range modulesByName {
		node := getMatchingUsesNodeFromMod(mod, name)
		if node != nil {
			if node.Kind() == "grouping" {
				return getMatchingUsesNode(node.NName())
			} 
			return node
		}
	}
	return nil
}

// Locate the node that includes the uses with the name. The search
// is recursive. This node could be included in any submodule or module
// associated with the submodule.
func getMatchingUsesNodeFromMod(mod *Module, name string) yang.Node {
	for _, sm := range mod.submodules {
		for _, g := range sm.module.Grouping {
			if node := getMatchingUsesNodeFromGrouping(g, name); node != nil {
				return node
			}
		}
		for _, c := range sm.module.Container {
			if node := getMatchingUsesNodeFromContainer(c, name); node != nil {
				return node
			}
		}
	}
	return nil
}

// "augment" is another way of inserting a node into the yang tree.
// We need to look for nodes considering augmented nodes too. This
// function aims at parsing through augments
func getNodeFromAugments(path, part string, node yang.Node) (*Module, yang.Node) {
	// Using the part, first locate module where to look for the
	// "augment" node
	ym := getMyYangModule(node)
	prefix := getPrefix(part)
	modname := getModuleNameFromPrefix(ym, prefix)
	mod, ok := modulesByName[modname]
	if !ok {
		return mod, nil
	}
	for _, sm := range mod.submodules {
		for _, aug := range sm.module.Augment {
			if aug.Name == path {
				//debuglog("getNodeFromAugments() - Located augment node %s", nodeContextStr(aug))
				node := getNodeFromAugment(aug, part)
				if node != nil {
					return mod, node
				}
			}
		}
	}
	return mod, nil
}
func getNodeFromAugment(aug *yang.Augment, part string) yang.Node {
	name := getName(part)
	needleaf := false

	// Process uses to locate the node indicated by part
	for _, u := range aug.Uses {
		_, node := getFromUsesNode(u, name, needleaf)
		if node != nil {
			return node
		}
	}
	for _, c := range aug.Container {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func traverse(path string, node yang.Node, needleaf bool) yang.Node {
	var root bool = false
	var accPath string
	var curr, next yang.Node

	debuglog("traverse(): Looking for %s needed by %s.%s", path, node.NName(), node.Kind())
	// yang supports quite a varied set of capabilities and we do not intend support them all.
	// We are going to trim a known form which has text between '[' and ']'. The removal
	// should not impact the traversal to locate the the required node for the purpose of 
	// learning the type.
	re := regexp.MustCompile("\\[.*\\]")
	path = re.ReplaceAllString(path, "")

	// Now break the path into individual components so that the
	// tree can be traversed for each individual component. This section
	// sets the starting point for the search. A relative path starts
	// with the current node as the starting point while a root path,
	// that starts with '/', sets the module as th starting point
	if strings.HasPrefix(path, "/") {
		root = true
	}
	parts := strings.Split(path, "/")

	// If the path is a root path, the first part is an empty part and is to be trimmed.
	// If the path is a root path, we must also set the starting point of the traversal
	// to the module in the first part of the root path. Else, it is a relative path and
	// the start is set to the 'node' passed to this function
	curr = node
	next = node
	if root {
		parts = parts[1:]
		prefix := getPrefix(parts[0])
		if prefix != "" {
			ymod := getImportedYangModuleByPrefix(getMyYangModule(node), prefix)
			if ymod == nil {
				errorlog("traverse(): Failed to find prefix=%s for %s.%s", prefix, node.NName(), node.Kind())
				return nil
			}
			next = ymod
		} else {
			next = getMyYangModule(node)
		}
	}

	// lets traverse and locate based on parts from the path provided
	for i, part := range parts {
		debuglog("traverse(): Locate \"%s\": iteration=%d/%d accpath=%s, path=%s",
				part, i, len(parts), accPath, path)
		switch part {
		case "..":
			next = next.ParentNode()
			/*
			if next == nil {
				errorlog("traverse(): Failed for .., i=%d, path=%s", i, path)
				return nil
			}
			*/
			if next.Kind() == "grouping" {
				// This is a grouping and is the higher level within a module. This
				// must be included somwhere for its instantiation and thus we should
				// look for it across the entire YANG specification
				debuglog("traverse(): Locate uses node %s in iteration = %d", next.NName(), i)
				next = getMatchingUsesNode(next.NName())
				if next == nil {
					errorlog("traverse(): Failed to find node that uses %s", curr.NName())
					break
				}
			}
		default:
			// From the "part", identify the module and name to search for.
			// The module is derived from the prefix in the "part".
			name := getName(part)
			next = getNextNodeByName(next, name, needleaf && i == (len(parts)-1))
		}
		// If we couldn't find the next node, we cannot continue further in the traversal
		// Let's return out of the function with failure
		if next == nil {
			errorlog("traverse(): Failed at index=%d name=%s, path=%s, accPath=%s, curr=%s.%s",
					i, part, path, accPath, curr.NName(), curr.Kind())
			return nil
		}
		debuglog("traverse(): Found index = %d/%d next node = %s", i, len(parts), next.NName())
		accPath = accPath + "/" + part
		curr = next
	}
	return curr
}

// This utility function creates directory if the directory doesn't exist
func ensureDirectory(path string) {
	dirName := filepath.Dir(path)
	if _, err := os.Stat(dirName); err != nil {
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			log.Fatalf("Error: %s", err.Error())
		}
	}
}

/*
func storeInPrefixModuleMap(m *yang.Module) {
	//debuglog("storing mod for prefix: %s", m.GetPrefix())
	p := getYangPrefix(m)
	prefixModulesMap[p] = append(prefixModulesMap[p], m)
}
func storeInGroupingMap(prefix string, g yang.Node) {
	//debuglog("storing grouping: %s", g.NName())
	groupingMap[prefix+":"+g.NName()] = g
}
func mergeModules() {
	for _, mods := range prefixModulesMap {
		// There may be multiple submodules for same prefix. Use the base module. and also merge all entries.
		m := mods[0]
		var entries []yang.Node
		for _, mod := range mods {
			if mod.Prefix != nil {
				m = mod
			}
			entries = append(entries, mod.Entries...)
		}
		p := getYangPrefix(m)
		m.Entries = entries
		prefixModuleMap[p] = m
	}
}
*/

// getAllNodesFromPath builds the list with all the nodes from yang in proper order so that
// the variables can be set properly later.
/*
func getAllNodesFromPath(path string) []*NodeInfo {
	if !strings.HasPrefix(path, "/") {
		log.Fatalf("Path has to start with root '/'")
	}
	parts := strings.Split(path, "/")
	parts = parts[1:]
	prefix := getPrefix(parts[0])

	m, ok := prefixModuleMap[prefix]
	if !ok {
		log.Fatalf("module not present for prefix %s", prefix)
	}

	var node *yang.Node
	var nodeInfo *NodeInfo
	nodes := []*NodeInfo{}
	mod := m
	for _, part := range parts {
		// Unions have '--' inside
		us := strings.Split(part, "--")

		elements := strings.Split(us[0], ":")
		if len(elements) != 2 {
			log.Fatalf("Path should be in prefix:entry format %s", part)
		}
		prefix := elements[0]
		entry := elements[1]
		if len(us) == 2 {
			entry += us[1]
		}
		// Try to dig inside node for next element if node is not nil
		if node != nil {
			parentNodeInfo := nodeInfo
			//debuglog("node check: ", (*node).NName(), entry)
			node, mod = getMatchedNode(*node, mod, prefix, entry)
			if node == nil {
				log.Fatalf("Did not find any node in yang for %s:%s", prefix, entry)
			}
			nodeInfo = &NodeInfo{node, mod, parentNodeInfo}
			nodes = append(nodes, nodeInfo)
			//debuglog("node added: ", (*node).NName())
		} else {
			// If node is nil that means it is the 1st node. Try to find it from grouping.
			for _, e := range m.Entries {
				node, mod = getMatchedNode(e, m, prefix, entry)
				if node != nil {
					nodeInfo = &NodeInfo{node, mod, nil}
					nodes = append(nodes, nodeInfo)
					//debuglog("node added: ", (*node).NName())
					break
				}
			}
			if node == nil {
				log.Fatalf("Did not find any node in yang for %s:%s", prefix, entry)
			}
		}
	}

	return nodes
}

func getMatchedNode(node yang.Node, m *yang.Module, prefix, entry string) (*yang.Node, *yang.Module) {
	switch node.Kind() {
	case "container":
		c, ok := node.(*yang.Container)
		if !ok {
			panic("Not a Container")
		}
		return getMatchedNodeInside(c, m, prefix, entry)
	case "list":
		l, ok := node.(*yang.List)
		if !ok {
			panic("Not a List")
		}
		return getMatchedNodeInside(l, m, prefix, entry)
	case "grouping":
		g, ok := node.(*yang.Grouping)
		if !ok {
			panic("Not Grouping.")
		}
		return getMatchedNodeInside(g, m, prefix, entry)
	}
	//debuglog("1 ", node.Kind(), node.NName())
	return nil, nil
}

func getMatchedNodeInside(c *yang.Container, m *yang.Module, prefix, entry string) (*yang.Node, *yang.Module) {
	// if input prefix and current mod prefix is not matching then search in augments
	if m.Prefix.Name != prefix {
		augMod, ok := prefixModuleMap[prefix]
		if !ok {
			log.Fatalf("module not present for prefix %s", prefix)
		}
		for _, e := range augMod.Entries {
			if e.NName() == entry {
				// Check that node is actually present in augment
				//if !CheckNodePresentInAugment(e, c, prefix) {
				//	log.Fatalf("Did not find any node in yang for %s:%s", prefix, entry)
				//}
				return &e, augMod
			}
		}
	}

	entries := c.GetEntries()
	for _, e := range entries {
		switch e.Kind() {
		case "container":
			c, ok := e.(*yang.Container)
			if !ok {
				panic("Not a Container")
			}
			if entry == c.Name {
				return &e, m
			}
		case "list":
			l, ok := e.(*yang.List)
			if !ok {
				panic("Not a List")
			}
			if entry == l.Name {
				return &e, m
			}
		case "uses":
			u, ok := e.(*yang.Uses)
			if !ok {
				panic("Not a Uses")
			}
			//debuglog("***", u.Name)
			if entry == u.Name {
				g, ok := groupingMap[prefix+":"+entry]
				if !ok {
					panic("Grouping not found")
				}
				newMod := m
				if gPrefix := getPrefix(g.NName()); gPrefix == prefix {
					newMod = prefixModuleMap[prefix]
				}
				return &g, newMod
			}
		case "leaf-list":
			l, ok := e.(*yang.LeafList)
			if !ok {
				panic("Not a LeafList")
			}
			if entry == l.Name {
				return &e, m
			}
		case "leaf":
			l, ok := e.(*yang.Leaf)
			if !ok {
				panic("Not a LeafList")
			}
			if entry == l.Name {
				return &e, m
			}
		}
		//debuglog("2 ", e.Kind(), e.NName())
	}

	return nil, nil
}

func CheckNodePresentInAugment(n yang.Node, c yang.Container, prefix string) bool {
	for _, e := range c.GetAugments() {
		augYangMod := getMyYangModule(e)
		aug, ok := e.(*yang.Augment)
		if !ok {
			panic("The node is not an Augment: " + nodeString(e))
		}
		switch {
		case len(aug.Uses) > 0:
			for _, u := range aug.Uses {
				if u.Name == n.NName() && augYangMod.Prefix.Name == prefix {
					return true
				}
			}
		case len(aug.Leaf) > 0:
			for _, l := range aug.Leaf {
				if l.Name == n.NName() && augYangMod.Prefix.Name == prefix {
					return true
				}
			}
		case len(aug.LeafList) > 0:
			for _, l := range aug.LeafList {
				if l.Name == n.NName() && augYangMod.Prefix.Name == prefix {
					return true
				}

			}
		case len(aug.Container) > 0:
			for _, c := range aug.Container {
				if c.Name == n.NName() && augYangMod.Prefix.Name == prefix {
					return true
				}
			}
		default:
			debuglog("ERROR: Augment case not supported yet:", nodeContextStr(aug))
		}
	}
	return false
}

// Keep a map with path and node data.
var Nodes map[string]*NodeData

// Here we will try to build a tree with all the paths combined. Index in nodesTree will resemble levels in path and
// each level can have mutiple shibling nodes.
var RootNodes []*NodeData

func initNodesTree() {
	RootNodes = []*NodeData{}
	Nodes = map[string]*NodeData{}
}

func updateNodesTreeFromPath(path string, structName string, elem *Element) {
	log.Println("Working on path: ", path)
	if !strings.HasPrefix(path, "/") {
		log.Fatalf("Path has to start with root '/'")
	}
	parts := strings.Split(path, "/")
	parts = parts[1:]
	prefix := getPrefix(parts[0])

	m, ok := prefixModuleMap[prefix]
	if !ok {
		log.Fatalf("module not present for prefix %s", prefix)
	}

	var node *yang.Node
	var nodeInfo *NodeData
	var varName string
	mod := m
	for idx, part := range parts {
		key := "<" + strconv.Itoa(idx) + ">" + part
		if n, ok := Nodes[key]; ok {
			node = n.node
			nodeInfo = n
			varName = n.varName
			mod = n.mod
			continue
		}

		// Unions have '--' inside
		us := strings.Split(part, "--")

		elements := strings.Split(us[0], ":")
		if len(elements) != 2 {
			log.Fatalf("Path should be in prefix:entry format %s", part)
		}
		prefix := elements[0]
		entry := elements[1]
		if len(us) == 2 {
			entry += "--" + us[1]
		}

		var varType, kind string

		// Try to dig inside node for next element if node is not nil
		if node != nil {
			parentNodeInfo := nodeInfo
			parentVarName := varName
			//debuglog("node check: ", (*node).NName(), entry)
			node, mod, varType, varName, kind = findMatchedNode(*node, mod, prefix, entry)
			if node == nil {
				log.Fatalf("1. Did not find any node in yang for %s:%s", prefix, entry)
			}
			if varName != "" {
				varName = parentVarName + "." + varName
			} else {
				varName = parentVarName
			}
			nodeInfo = &NodeData{node, mod, varType, varName, kind,
				structName, elem, parentNodeInfo, []interface{}{}}
			parentNodeInfo.children = append(parentNodeInfo.children, nodeInfo)
			Nodes[key] = nodeInfo
			//debuglog("node added: ", (*node).NName())
		} else {
			// If node is nil that means it is the 1st node. Try to find it from grouping.
			for _, e := range m.Entries {
				node, mod, varType, varName, kind = findMatchedNode(e, m, prefix, entry)
				if node != nil {
					nodeInfo = &NodeData{node, mod, varType, varName, kind,
						structName, elem, nil, []interface{}{}}
					//debuglog("node added: ", (*node).NName())
					// This is a root nodeInfo
					RootNodes = append(RootNodes, nodeInfo)
					Nodes[key] = nodeInfo
					break
				}
			}
			if node == nil {
				log.Fatalf("2. Did not find any node in yang for %s:%s", prefix, entry)
			}
		}
	}
}

func findMatchedNode(node yang.Node, m *yang.Module, prefix, entry string) (*yang.Node, *yang.Module,
	string, string, string) {
	switch node.Kind() {
	case "container":
		c, ok := node.(*yang.Container)
		if !ok {
			panic("Not a Container")
		}
		return findMatchedNodeInside(c, m, prefix, entry)
	case "list":
		l, ok := node.(*yang.List)
		if !ok {
			panic("Not a List")
		}
		return findMatchedNodeInside(l, m, prefix, entry)
	case "grouping":
		g, ok := node.(*yang.Grouping)
		if !ok {
			panic("Not Grouping.")
		}
		return findMatchedNodeInside(g, m, prefix, entry)
	case "leaf":
		// If the path for a member variable is going beyond leaf that means it could be a union.
		// For now the support is upto union only.
		return findMatchedNodeForTypedef(prefix, entry)
	}
	//debuglog("1 ", node.Kind(), node.NName())
	return nil, nil, "", "", ""
}

func findMatchedNodeInside(c yang.Container, m *yang.Module, prefix, entry string) (*yang.Node,
	*yang.Module, string, string, string) {
	// if input prefix and current mod prefix is not matching then search in augments
	if m.Prefix.Name != prefix {
		augMod, ok := prefixModuleMap[prefix]
		if !ok {
			log.Fatalf("module not present for prefix %s", prefix)
		}
		for _, e := range augMod.Entries {
			if e.NName() == entry {
				// Check that node is actually present in augment
				//if !CheckNodePresentInAugment(e, c, prefix) {
				//	log.Fatalf("Did not find any node in yang for %s:%s", prefix, entry)
				//}
				return &e, augMod, "", "", e.Kind() // This is embedded struct hence no type and name for variable
			}
		}
		return nil, nil, "", "", ""
	}

	entries := c.GetEntries()
	for _, e := range entries {
		fn := e.NName()
		//debuglog("1 ", e.Kind(), e.NName())
		switch e.Kind() {
		case "container":
			c, ok := e.(*yang.Container)
			if !ok {
				panic("Not a Container")
			}
			if entry == c.Name {
				varType := fmt.Sprintf("%s_cont", genTN(m, c.NName()))
				varName := genFN(fn)
				return &e, m, varType, varName, e.Kind()
			}
		case "list":
			l, ok := e.(*yang.List)
			if !ok {
				panic("Not a List")
			}
			if entry == l.Name {
				tn := l.NName()
				varType := fmt.Sprintf("%s", genTN(m, tn))
				varName := genFN(fn) + "[0]"
				return &e, m, varType, varName, e.Kind()
			}
		case "uses":
			u, ok := e.(*yang.Uses)
			if !ok {
				panic("Not a Uses")
			}
			//debuglog("***", u.Name)
			if entry == u.Name {
				g, ok := groupingMap[prefix+":"+u.Name]
				if !ok {
					panic("Grouping not found")
				}
				newMod := m
				if gPrefix := getPrefix(g.NName()); gPrefix == prefix {
					newMod = prefixModuleMap[prefix]
				}
				return &g, newMod, "", "", e.Kind() // This is embedded struct hence no type and name for variable
			}
		case "leaf-list":
			l, ok := e.(*yang.LeafList)
			if !ok {
				panic("Not a LeafList")
			}
			if entry == l.Name {
				tn := l.NName()
				varType := fmt.Sprintf("%s[0]", genTN(m, tn))
				varName := genFN(fn)
				return &e, m, varType, varName, e.Kind()
			}
		case "leaf":
			l, ok := e.(*yang.Leaf)
			if !ok {
				panic("Not a Leaf")
			}
			if entry == l.Name {
				varType := getTypeName(m, l.Type)
				varName := genFN(fn)
				return &e, m, varType, varName, e.Kind()
			}
		}
	}

	return nil, nil, "", "", ""
}

func findMatchedNodeForTypedef(prefix, entry string) (*yang.Node,
	*yang.Module, string, string, string) {
	mod, ok := prefixModuleMap[prefix]
	if !ok {
		log.Fatalf("module not present for prefix %s", prefix)
	}

	tmp := strings.Split(entry, "--")
	if len(tmp) != 2 {
		log.Fatalf("union names are not proper. typename and union member name must be separated by --")
	}
	typeName := tmp[0]
	unionMemberName := tmp[1]
	//	fmt.Printf("findMatchedNodeForTypedef: %s", entry)
	for _, e := range mod.Entries {
		if e.Kind() == "typedef" && e.NName() == typeName {
			t, ok := e.(*yang.Typedef)
			if !ok {
				panic("Not a Typedef")
			}
			return findMatchedNodeForType(t.Type, mod, unionMemberName)
		}
	}
	return nil, nil, "", "", ""
}

func findMatchedNodeForType(n yang.Node, mod *yang.Module, unionMemberName string) (*yang.Node, *yang.Module, string, string, string) {
	t, ok := n.(*yang.Type)
	if !ok {
		panic("Not a Type")
	}

	switch t.Name {
	case "union":
		for id, it := range t.Type {
			if it.Name == unionMemberName {
				varName := fmt.Sprintf("%s_%d", genFN(it.Name), id)
				varType := getTypeName(mod, it)
				return &n, mod, varType, varName, "union"
			}
		}
	default:
		log.Fatalf("Does not support other types other than union")
	}
	return nil, nil, "", "", ""
}

func addPrefixToDatatypeIfRequired(datatype string, prefix string) string {
	switch datatype {
	case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "int", "uint", "rune", "byte",
		"float32", "float64", "string", "bool":
		return datatype
	default:
		return prefix + "." + datatype
	}
}
*/
