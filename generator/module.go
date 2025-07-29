package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

// Moudle structure design ************************
// Module maps to "module" of yang specification. As module may contain
// submodules, the design choice is to have the basic module itself embedded
// inside a module as a submodule. The Module struct is a shell which at 
// a minimum has one module in it even if it doesn't contain any submodules
// in the yang schema

// initfunc field has a list of statements that should be written to the
// init() function of the module in the generated code TODO review
type ModuleType int

const (
	TypeModule ModuleType = iota
	TypeSubModule
)

type SubModule struct {
	name      string
	mod       *Module
	module    *yang.Module
	mtype     ModuleType
	namespace string
	prefix    string
	initfunc  []string
}

var submodToMod = map[string]string{}

// Each module as in YANG specification has the following: a name, 
// a prefix (a short form for reference), a namespace, // etc.
// * field "identities" is introduced to capture the result of 
// preprocessing of identities as consolidation is needed for 
// generating the code related to identifies.
// * field submodules includes all submodules of the module of which
// the module itself is one. This is because it makes the logic
// more maintainable and easy to write and read.
type Module struct {
	name                    string
	prefix                  string
	namespace               string
	imports                 map[string]string
	identities              map[string]*yang.Identity
	submodules              map[string]*SubModule
}

// Constructor for structure Module
func NewModule(m *yang.Module) *Module {
	mod := &Module{}
	mod.name = m.NName()
	mod.prefix = m.Prefix.Name
	mod.imports = make(map[string]string)
	mod.identities = make(map[string]*yang.Identity)
	mod.submodules = make(map[string]*SubModule)
	mod.namespace = m.Namespace.Name
	mod.prefix = m.Prefix.Name

	// Add the module also as a submodule which is used
	// for any processing related to generation of code
	submod := &SubModule{}
	submod.mod = mod
	submod.mtype = TypeModule
	submod.module = m
	submod.name = m.Name
	mod.submodules[submod.name] = submod
	submodToMod[submod.name] = submod.name
	return mod
}

func printModule(m *Module) {
	fmt.Println("Module:", m.name)
	indent := 0
	for _, sm := range m.submodules {
		mod := sm.module	
		printYangModule(mod, indent + 1)
	}
}


// Maps of modules by the prefixes given by the modules instead of
// prefixes used by different modules for importing. This is a global
// list
var modulesByPrefix = map[string]*Module{}
var modulesByName = map[string]*Module{}

// Manage identities for code generation. If an identity is referred
// to by another identity, type definition/marshal/unmarshal/etc. are
// needed for the base identity. The preprocessing is expected to fill
// the map with the identities for which code must be generated

// add an identity for which code must be generated
func (m *Module) addBaseIdentity(id *yang.Identity) {
	m.identities[id.Name] = id
}

// get an identity for which code must be generated
func (m *Module) getBaseIdentity(name string) *yang.Identity {
	if id, ok := m.identities[name]; ok {
		return id
	}
	return nil
}

// Add a module to the map of modules maintained based on prefix and name
// Populate the prefix to the module mapping. The prefix used to represent a
// module in other modules doesn't have to match the prefix used in the module
func addModule(mod *yang.Module) {
	m := NewModule(mod)
	if tm, ok := modulesByName[m.name]; ok {
		errorlog("Module by name already exists: %s", tm.name)
		return
	}
	modulesByName[m.name] = m
	if tm, ok := modulesByPrefix[m.prefix]; ok {
		errorlog("Module by prefix already exists: %s", tm.prefix)
		return
	}
	modulesByPrefix[m.prefix] = m
}

// Adds a submodule to the map within a module. For any prefix/namespace
// based search, the traversal should include submodule too.
func addSubModule(m *yang.Module) {
	lm := &SubModule{}
	lm.name = m.Name
	lm.mtype = TypeSubModule
	lm.module = m
	modname := m.BelongsTo.Name
	if mod, ok := modulesByName[modname]; ok {
		mod.submodules[lm.name] = lm
		lm.mod = mod
		lm.prefix = mod.prefix
		lm.namespace = mod.namespace
		submodToMod[lm.name] = mod.name
	} else {
		panic("Failed to add submodule")
	}
}
func getSubModule(name string) *SubModule {
	modname, ok := submodToMod[name]
	if !ok {
		return nil
	}
	mod, ok := modulesByName[modname]
	if !ok {
		return nil
	}
	submod, ok := mod.submodules[name]
	if !ok {
		return nil
	}
	return submod
}

// Locate a module by the prefix and return it
func getModuleByPrefix(pre string) *Module {
	if mod, ok := modulesByPrefix[pre]; ok {
		return mod
	}
	return nil
}

// Currently, the preprocessing involves only the identities. We will
// need to expand this as we identify more use cases such as "uses",
// especially for generating the namespace related text into the type
// definitions
func (m *Module) preprocessModule() {
	debuglog("Preprocessing module: %s", m.name)
	m.preprocessIdentities()
	m.preprocessAugments()
}

// This function generates the common initial part of the go file for a
// module.
func fileHeader(mod *Module, submod *SubModule, w io.Writer, keepXmlID bool) {
	modname := genFN(mod.name)
	submodname := genFN(submod.name)
	// Generic header of the file with imports and package
	// In future, we should take package name as an attribute
	fmt.Fprintf(w, "package yang\n")
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, "\t\"fmt\"\n")
	fmt.Fprintf(w, "\t\"strings\"\n")
	fmt.Fprintf(w, "\t\"regexp\"\n")
	fmt.Fprintf(w, "\t\"strconv\"\n")
	fmt.Fprintf(w, "\t\"math\"\n")
	fmt.Fprintf(w, "\t\"encoding/base64\"\n")
	if keepXmlID {
		fmt.Fprintf(w, "\tnc \"toradapter/lib/encoding/nc\"\n")
	}
	fmt.Fprintf(w, ")\n")
	// Add comments to the file that provide the information about the
	// source file that was used to generate the code
	addFileComments(w, submod.module)

	// These declarations apply only to main module. The submodules
	// use the definitions of the main module
	if submod.mtype == TypeModule {
		fmt.Fprintf(w, "var %s_ns = \"%s\"\n", modname, mod.namespace)
		fmt.Fprintf(w, "var %s_prefix = \"%s\"\n", modname, mod.prefix)
	}
	// Revision is needed independently for both submodules and main module
	// and must be generated outside the earlier check
	if submod.mtype == TypeModule {
		revision := submod.module.Revision[0].Name
		fmt.Fprintf(w, "var %s_capability = \"%s?module=%s&revision=%s\"\n", modname, mod.namespace, mod.name, revision)
	} else {
		//fmt.Fprintf(w, "var %s_capability = \"%s?\"\n", submodname, mod.namespace)
	}
	fmt.Fprintf(w, "var %s_revision = \"%s\"\n", submodname, submod.module.Revision[0].Name)
	// Generate the dummy usage for all the common import packages so that
	// we don't have to carefully identify which of them to be included
	fmt.Fprintf(w, "\n//-----------------------------------------------------\n")
	fmt.Fprintf(w, "//Dummy code to avoid careful insertion of imports\n")
	fmt.Fprintf(w, "var %s_strings = strings.HasSuffix(\"dummy\", \"d\")\n", submodname)
	fmt.Fprintf(w, "var %s_re = regexp.MustCompile(\"dummy\")\n", submodname)
	fmt.Fprintf(w, "var %s_xy = strconv.FormatInt(10,10)\n", submodname)
	fmt.Fprintf(w, "var %s_math = math.Abs(10.0)\n", submodname)
	fmt.Fprintf(w, "var %s_err = fmt.Errorf(\"dummy\")\n", submodname)
	fmt.Fprintf(w, "var %s_base64 = base64.StdEncoding\n", submodname)
	fmt.Fprintf(w, "//-----------------------------------------------------\n\n")
}
func addFileComments(w io.Writer, ymod *yang.Module) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Module Name:\n")
	s := indentString(ymod.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	s = indentString(ymod.Description.Name)
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprintf(w, "// Revisions:\n")
	for _, r := range ymod.Revision {
		s = indentString("\n" + r.Description.Name)
		s = indentString(r.Name + s)
		s = commentString(s)
		fmt.Fprint(w, s)
	}
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

// Process the main module and its submodules
func processModule(mod *Module, outdir string) {
	for _, sm := range mod.submodules {
		fmt.Println("***********Processing module", sm.module.NName(), "...")
		processSubModule(mod, sm, outdir)
		storeInPrefixModuleMap(sm.module)
	}
}

// Process a module to do the following:
// - generate file header
// - process all the entries of module
// - generate the init() function
func processSubModule(mod *Module, submod *SubModule, outdir string) {
	// prepare the essentials for the module
	m := submod.module
	inpath := m.Source.Location()
	_, file := path.Split(inpath)
	mainname := strings.Split(file, ".yang")
	outpath := outdir + "/yang-go/" + mainname[0] + ".go"

	// If outpath is not present then create it.
	ensureDirectory(outpath)

	debuglog("Processing file %s%s", mainname[0], "...")
	w, err := os.OpenFile(outpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		errorlog("unable to open file %s", err.Error())
	}

	groupingNames := map[string]struct{}{}
	for _, u := range submod.module.Uses {
		groupingNames[u.Name] = struct{}{}
	}
	// TODO
	//keepXmlID := checkIfNcImportRequired(submod.module.Entries, groupingNames)
	keepXmlID := true

	// Create file header
	fileHeader(mod, submod, w, keepXmlID)

	//entries := mergeAugmentsWithSamePath(submod.module.Entries)
	// process the entries of the module
	for _, i := range submod.module.Identity {
		processIdentity(w, submod, m, i)
	}
	for _, g := range submod.module.Grouping {
		processGrouping(w, submod, m, g, keepXmlID)
	}
	for _, t := range submod.module.Typedef {
		processTypedef(w, submod, m, t)
	}

	// generate the init() function
	fmt.Fprintf(w, "func init() {\n")
	for _, s := range submod.initfunc {
		fmt.Fprintf(w, "\t%s", s)
	}
	fmt.Fprintf(w, "}\n")
	w.Close()
}

// To use XMLname we need nc module to be imported. This function check whether we need it or not.
func checkIfNcImportRequired(entries []yang.Node, groupingNames map[string]struct{}) bool {
	for _, e := range entries {
		switch e.Kind() {
		case "grouping":
			g, ok := e.(*yang.Grouping)
			if !ok {
				panic("Not Grouping.")
			}
			if _, ok := groupingNames[g.Name]; ok {
				return true
			}
		}
	}
	return false
}

// There might be multiple augments with same path.
// Merge all augments with same path so that only one struct can be generated.
func mergeAugmentsWithSamePath(entries []yang.Node) []yang.Node {
	newEntries := []yang.Node{}
	augments := map[string]*yang.Augment{}

	for _, e := range entries {
		if e.Kind() == "augment" {
			aug, ok := e.(*yang.Augment)
			if !ok {
				panic("Not an Augment")
			}
			if a, found := augments[aug.NName()]; found {
				// Merge the contents.
				a.LeafList = append(a.LeafList, aug.LeafList...)
				a.Leaf = append(a.Leaf, aug.Leaf...)
				a.Container = append(a.Container, aug.Container...)
				a.Uses = append(a.Uses, aug.Uses...)
				a.List = append(a.List, aug.List...)
				// So on. As we keep supporting more we will add here.
			} else {
				augments[aug.NName()] = aug
			}
		} else {
			newEntries = append(newEntries, e)
		}
	}
	for _, aug := range augments {
		newEntries = append(newEntries, yang.Node(aug))
	}

	return newEntries
}

// One of the utility functions that help traversal across the YANG specification
func getGroupingFromMod(mod *Module, name string) *yang.Grouping {
	prefix := getPrefix(name)
	gname := getName(name)
	if prefix != "" {
		mod = getModuleByPrefix(prefix)
	}
	if mod == nil {
		return nil
	}
	for _, sm := range mod.submodules {
		ymod := sm.module
		for _, g := range ymod.Grouping {
			if g.NName() == gname {
				return g
			}
		}
	}
	return nil
}
