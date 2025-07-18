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

var prefixModulesMap = map[string][]*yang.Module{}
var prefixModuleMap = map[string]*yang.Module{}
var groupingMap = map[string]yang.Node{}

func getEntries(c *yang.Container) []yang.Node {
	e := make([]yang.Node, 0, 1000)
	for _, l := range c.Leaf {
		e = append(e, l)
	}
	return e
}

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
}

func printModule(m *Module) {
	fmt.Println("Module:", m.name)
	indent := 0
	for _, sm := range m.submodules {
		mod := sm.module	
		printYangModule(mod, indent + 1)
	}
}

// Function that emits the name of the node in a short form
// to include in debug statements rather efficiently
func nodeString(node yang.Node) string {
	return node.Kind() + ":" + node.NName()
}

// Function to fully describe the node
func nodeStringFull(node yang.Node) string {
	var mname string
	ymod := getMyYangModule(node)
	if ymod.BelongsTo != nil {
		mname = ymod.BelongsTo.Name
	} else {
		mname = ymod.Name
	}

	return node.Kind() + ":" + node.NName() + "@ module:" + mname
}

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
// tree.
func getMyYangModule(node yang.Node) *yang.Module {
	n := node
	for n.ParentNode() != nil {
		n = n.ParentNode()
	}
	ymod, ok := n.(*yang.Module)
	if !ok {
		panic("Module not found")
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
		panic("getMyModule() - module is nil")
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
			mod, ok := modulesByName[i.Name]
			if ok {
				return mod
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

// Traverse a "Uses" for locating the node. The traversal could be
// recursive. Such a traversal can switch the module and thus we
// need to return both module and node to further the traversal thereon
func getFromUsesNode(n yang.Node, name string, needleaf bool) (*Module, yang.Node) {
	//fmt.Println("getFromUsesNode() - Looking for", name, "in", nodeStringFull(n))
	u, ok := n.(*yang.Uses)
	if !ok {
		panic("Not a Uses node")
	}

	// The uses may point to a different module. That is determined
	// based on the prefix qualification in the "uses" name
	prefix := getPrefix(u.Name)
	gname := getName(u.Name)
	m := getMyYangModule(u)
	modname := getModuleNameFromPrefix(m, prefix)
	mod, ok := modulesByName[modname]
	if !ok {
		// fmt.Println("getFromUsesNode(): Module not found for prefix " + prefix + " in module " + m.Name)
		return nil, nil
	}
	node := getNodeFromModule(mod, gname, false)
	if node == nil {
		return mod, nil
	}
	if c, ok := node.(*yang.Container); ok {
		return getNodeFromContainer(mod, c, name, needleaf)
	}
	return mod, nil
}

// Get node from a container. This function does not know the
// type of container and just examines the entries in the container
// to locate the node by name. It may even have to traverse nodes
// of another node included using "uses" of yang.
func getNodeFromContainer(mod *Module, c *yang.Container, name string, needleaf bool) (retm *Module, retn yang.Node) {
	//fmt.Println("getNodeFromContainer() - looking for",  name, "in", nodeStringFull(c))
	defer func() {
		if retn == nil {
			//fmt.Println("getNodeFromContainer() - return nil node")
		} else {
			//fmt.Println("getNodeFromContainer() - return with module =", retm.name, "node =", nodeStringFull(retn))
		}
	}()
	//for _, e := range c.GetEntries() {
	for _, e := range getEntries(c) {
		if e.Kind() == "uses" {
			if mod, node := getFromUsesNode(e, name, needleaf); node != nil {
				return mod, node
			}
		}
		if e.NName() == name {
			if needleaf {
				if e.Kind() == "leaf" {
					return mod, e
				}
			} else {
				if e.Kind() == "grouping" || e.Kind() == "container" ||
					e.Kind() == "augment" || e.Kind() == "list" ||
					e.Kind() == "leaf-list" {
					return mod, e
				}
			}
		}
	}
	/*
		for _, a := range c.GetAugments() {
			aug, ok := a.(*yang.Augment)
			if !ok {
				panic("The node from GetAugments() is not an augment: " + nodeStringFull(a))
			}
			if e := getNodeFromAugment(aug, name); e != nil {
				nmod := getMyModule(e)
				return nmod, e
			}
		}
	*/
	return mod, nil
}

// This utility locates the next node by name and returns it
// If the current node includes a uses, it must traverse the
// entries within the type pointed by uses. Thus, the traversal
// can be recursive.
func getNodeByName(mod *Module, curr yang.Node, name string, leaf bool) (*Module, yang.Node) {
	// fmt.Println("getNodeByName() - Looking for", name, "in curr node", curr.NName())
	if curr != nil {
		// Since current node is passed, the searching for the next node is performed
		// within the current node. The node passed must be a container which has a
		// list of entries to search from
		c, ok := curr.(*yang.Container)
		if ok {
			return getNodeFromContainer(mod, c, name, leaf)
		} else {
			fmt.Println("ERROR: Not a container:", nodeStringFull(curr))
			return mod, nil
		}
	} else {
		for _, sm := range mod.submodules {
			m := sm.module
			// TODO Changed to module from container. Verify
			nmod, node := getNodeByName(mod, m, name, leaf)
			if node != nil {
				return nmod, node
			}
		}
	}
	return mod, nil
}

// Locate the node from the entire set of modules from the uses.
func getNodeWithUsesGlobal(modname string, name string) (*Module, yang.Node) {
	//fmt.Println("getNodeWithUsesGlobal() - Searching for", name)
	for _, mod := range modulesByName {
		_, node := getNodeWithUsesFromMod(mod, modname, name)
		if node != nil {
			return mod, node
		}
	}
	return nil, nil
}

// Locate the node that includes the uses with the name. The search
// is recursive. This node could be included in any submodule or module
// associated with the submodule.
func getNodeWithUsesFromMod(mod *Module, modname string, name string) (*Module, yang.Node) {
	for _, sm := range mod.submodules {
		// fmt.Println("getNodeWithUsesFromMod() - Searching for", name, "in", sm.module.Name)
		if node := getNodeWithUses(sm.module, modname, name); node != nil {
			//fmt.Println("getNodeWithUsesFromMod() - Found node", nodeStringFull(node))
			return mod, node
		}
	}
	return nil, nil
}
func getNodeWithUses(n yang.Node, modname string, name string) yang.Node {
	if m, ok := n.(*yang.Module); ok {
		for _, lc := range m.Container {
			if node := getNodeWithUses(lc, modname, name); node != nil {
				return node
			}
		}
	}
	if lc, ok := n.(*yang.Container); ok {
		for _, e := range lc.Uses {
			uname := getName(e.NName())
			upre := getPrefix(e.NName())
			ymod := getMyYangModule(e)
			umodname := getModuleNameFromPrefix(ymod, upre)
			if uname == name && modname == umodname {
					return e
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
				//fmt.Println("getNodeFromAugments() - Located augment node", nodeStringFull(aug))
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

// This utility performs traversal along the tree which
// includes the set of modules processed and locates the
// node. The path can be relative or absolute
func traverse(path string, node yang.Node, needleaf bool) yang.Node {
	var root bool = false
	var mod, nmod *Module
	var accPath string

	// Complete some initialization
	mod = getMyModule(node)
	ymod := getMyYangModule(node)
	curr := node
	next := node

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

	// If the path is a root path, the first part is an empty
	// part and is to be trimmed
	if root {
		parts = parts[1:]
		prefix := getPrefix(parts[0])
		if prefix != "" {
			mod = getImportedModuleByPrefix(ymod, prefix)
			if mod == nil {
				return nil
			}
		}
		next = nil
	}

	// lets traverse starting with the node and locate based on parts
	// collected earlier
	nmod = mod
	for i, part := range parts {
		// From the "part", identify the module and name to search for.
		// The module is derived from the prefix in the "part".
		pre := getPrefix(part)
		if pre != "" {
			ym := getMyYangModule(node)
			nmod := getImportedModuleByPrefix(ym, pre)
			if nmod == nil {
				//fmt.Println("ERROR: traverse() - module couldn't be found for prefix =", pre)
				return nil
			}
			if nmod != mod {
				next = nil
			}
		}
		name := getName(part)
		//fmt.Println("Accumulated path =", accPath)
		switch name {
		case "..":
			for next.Kind() == "grouping" {
				nmod, next = getNodeWithUsesGlobal(nmod.name, next.NName())
				if next == nil {
					break
				}
				next = next.ParentNode()
				// fmt.Println("Index",i,"/",len(parts), "Located node:", nodeStringFull(next))
			}
			if next == nil {
				break
			}
			next = next.ParentNode()
			//fmt.Println("Index",i,"/",len(parts), "Leaving with node: ", nodeStringFull(next))
		default:
			// Locate the node in the current container
			nmod, next = getNodeByName(nmod, next, name, needleaf && i == (len(parts)-1))
			if next == nil {
				// We didn't find it here so look for other places the same container
				// is included using "uses"
				next = curr
				for next.Kind() == "grouping" {
					nmod, next = getNodeWithUsesGlobal(nmod.name, next.NName())
					if next == nil {
						break
					}
					next = next.ParentNode()
					//fmt.Println("Index",i,"/",len(parts), "Located node:", nodeStringFull(next))
				}
				if next == nil {
					break
				}
				// fmt.Println("Index",i,"/",len(parts), "looking for", name, "@", nodeStringFull(next))
				nmod, next = getNodeByName(nmod, next, name, needleaf && i == (len(parts)-1))
			}
			// Look for augments to locate the node
			if next == nil {
				nmod, next = getNodeFromAugments(accPath, part, node)
			}
			if next != nil {
				//fmt.Println("Index",i,"/",len(parts), "Leaving with node", nodeStringFull(next))
			}
		}
		if next == nil {
			fmt.Println("ERROR: traversal failed at index =", i, ", path =", path)
			return nil
		} else {
			// fmt.Println("Trace: index =", i, "/", len(parts), "next node =", nodeStringFull(next))
		}
		accPath = accPath + "/" + part
		curr = next
		mod = nmod
	}
	return curr
}

// Find a leaf for a leafref which may even be recursive. Traverse
// till you find a node that isn't leafref
func getLeaf(path string, m *yang.Module, n yang.Node) *yang.Leaf {
	// Traverse the tree to fetch the node pointed to by the leafref.
	//fmt.Println("Trace: ************* fetching leaf for path =", path, "node =", n.NName())
	node := traverse(path, n, true)
	if node == nil {
		return nil
	}
	l, ok := node.(*yang.Leaf)
	if !ok {
		fmt.Println("ERROR: Found", nodeStringFull(node), "for path", path)
	}

	// Here we handle the case of a leafref pointing to another leafref and
	// so on. This lower portion addresses recursive traversal to locate the
	// final leaf of interest
	if l.Type.Name == "leafref" {
		// fmt.Println("Trace: ************** found leafref path =", l.Type.Path.Name)
		mod := getMyModule(l)
		if mod == nil {
			panic("Module not found for " + path + "leaf =" + l.Name)
		}
		return getLeaf(l.Type.Path.Name, m, l)
	}
	return l
}

func ensureDirectory(path string) {
	dirName := filepath.Dir(path)
	if _, err := os.Stat(dirName); err != nil {
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			log.Fatalf("Error: %s", err.Error())
		}
	}
}

// To keep the fmt import always in :)
func dummy() {
	fmt.Println("dummy")
}

func storeInPrefixModuleMap(m *yang.Module) {
	//fmt.Println("storing mod for prefix:", m.GetPrefix())
	p := getYangPrefix(m)
	prefixModulesMap[p] = append(prefixModulesMap[p], m)
}
func storeInGroupingMap(prefix string, g yang.Node) {
	//fmt.Println("storing grouping:", g.NName())
	groupingMap[prefix+":"+g.NName()] = g
}
/*
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
			//fmt.Println("node check: ", (*node).NName(), entry)
			node, mod = getMatchedNode(*node, mod, prefix, entry)
			if node == nil {
				log.Fatalf("Did not find any node in yang for %s:%s", prefix, entry)
			}
			nodeInfo = &NodeInfo{node, mod, parentNodeInfo}
			nodes = append(nodes, nodeInfo)
			//fmt.Println("node added: ", (*node).NName())
		} else {
			// If node is nil that means it is the 1st node. Try to find it from grouping.
			for _, e := range m.Entries {
				node, mod = getMatchedNode(e, m, prefix, entry)
				if node != nil {
					nodeInfo = &NodeInfo{node, mod, nil}
					nodes = append(nodes, nodeInfo)
					//fmt.Println("node added: ", (*node).NName())
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
	//fmt.Println("1 ", node.Kind(), node.NName())
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
			//fmt.Println("***", u.Name)
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
		//fmt.Println("2 ", e.Kind(), e.NName())
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
			fmt.Println("ERROR: Augment case not supported yet:", nodeStringFull(aug))
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
			//fmt.Println("node check: ", (*node).NName(), entry)
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
			//fmt.Println("node added: ", (*node).NName())
		} else {
			// If node is nil that means it is the 1st node. Try to find it from grouping.
			for _, e := range m.Entries {
				node, mod, varType, varName, kind = findMatchedNode(e, m, prefix, entry)
				if node != nil {
					nodeInfo = &NodeData{node, mod, varType, varName, kind,
						structName, elem, nil, []interface{}{}}
					//fmt.Println("node added: ", (*node).NName())
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
	//fmt.Println("1 ", node.Kind(), node.NName())
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
		//fmt.Println("1 ", e.Kind(), e.NName())
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
			//fmt.Println("***", u.Name)
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
