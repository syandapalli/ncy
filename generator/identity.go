package main

import (
	"fmt"
	"io"

	"github.com/openconfig/goyang/pkg/yang"
)

func addIdentityComment(w io.Writer, i *yang.Identity) {
	fmt.Fprintln(w, "//------------------------------------------------------------")
	fmt.Fprint(w, "//  Name:\n")
	s := indentString("identity: " + i.NName())
	s = commentString(s)
	fmt.Fprint(w, s)
	fmt.Fprint(w, "//  Description:\n")
	if i.Description != nil {
		s = indentString(i.Description.Name)
		s = commentString(s)
		fmt.Fprint(w, s)
	}
	fmt.Fprintln(w, "//-------------------------------------------------------------")
}

// If the identity has no base, its definition is created within the processing
// of the module. The derived ones are generated within the module where they
// are defined and are done only when another identity uses this as the base
func generateTypeDef(w io.Writer, m *yang.Module, id *yang.Identity) {
	var mname string
	if m.BelongsTo != nil {
		mname = m.BelongsTo.Name
	} else {
		mname = m.Name
	}

	// This is base of a an identity branch. Lets create a type definition
	// for it and also the maps to store the nodes that are part of the branch
	addIdentityComment(w, id)

	tn := genTN(m, id.Name) + "_id"
	fmt.Fprintf(w, "type %s string\n", tn)
	fmt.Fprintf(w, "var %s_prefix_map = map[string]string{}\n", tn)
	fmt.Fprintf(w, "var %s_ns_map = map[string]string{}\n", tn)

	// Write the Marshal function
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	fmt.Fprintf(w, "\tprefix, ok := %s_prefix_map[string(x)]\n", tn)
	fmt.Fprintf(w, "\tif ok {\n")
	fmt.Fprintf(w, "\t\tprefix = prefix + \":\"\n")
	fmt.Fprintf(w, "\t}\n")
	//fmt.Fprintf(w, "\tif lns, ok := %s_ns_map[string(x)]; !ok {\n", tn)
	//fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid %s = %%s\", string(x))\n", tn)
	//fmt.Fprintf(w, "\t} else if ns != lns {\n")
	//fmt.Fprintf(w, "\t\treturn []byte(prefix + string(x)), nil\n")
	fmt.Fprintf(w, "\treturn []byte(prefix + string(x)), nil\n")
	//fmt.Fprintf(w, "\t}\n")
	//fmt.Fprintf(w, "\treturn []byte(x), nil\n")
	fmt.Fprintf(w, "}\n")
	// Write the unmarshal function
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tvar name string\n")
	fmt.Fprintf(w, "\ts := string(b)\n")
	fmt.Fprintf(w, "\tparts := strings.Split(s, \":\")\n")
	fmt.Fprintf(w, "\tif len(parts) == 1 {\n")
	fmt.Fprintf(w, "\t\tname = parts[0]\n")
	fmt.Fprintf(w, "\t} else {\n")
	fmt.Fprintf(w, "\t\tname = parts[1]\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\tif _, ok := %s_ns_map[name]; !ok {\n", tn)
	fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid %s : %%s\", s)\n", tn)
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\t*x = %s(name)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
	// Write to runtime ns function
	fmt.Fprintf(w, "func (x %s)RuntimeNs() string {\n", tn)
	fmt.Fprintf(w, "\tif ns, ok := %s_ns_map[string(x)]; ok {\n", tn)
	//fmt.Fprintf(w, "\t\tif ns != %s_ns {\n", genFN(mname))
	//fmt.Fprintf(w, "\t\t\tif prefix, ok := %s_prefix_map[string(x)]; ok {\n", tn)
	//fmt.Fprintf(w, "\t\t\t\treturn prefix+\"!\"+ns\n")
	//fmt.Fprintf(w, "\t\t\t} else {\n")
	//fmt.Fprintf(w, "\t\t\t\treturn ns\n")
	//fmt.Fprintf(w, "\t\t\t}\n")
	//fmt.Fprintf(w, "\t\t}\n")
	fmt.Fprintf(w, "\t\tif prefix, ok := %s_prefix_map[string(x)]; ok {\n", tn)
	fmt.Fprintf(w, "\t\t\treturn prefix+\"!\"+ns\n")
	fmt.Fprintf(w, "\t\t} else {\n")
	fmt.Fprintf(w, "\t\t\treturn ns\n")
	fmt.Fprintf(w, "\t\t}\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\treturn %s_ns\n", genFN(mname))
	fmt.Fprintf(w, "}\n")
	// Just a seperator for readability
	fmt.Fprintf(w, "\n")
	return
}

func locateBase(m *yang.Module, input *yang.Identity) (*yang.Module, *yang.Identity) {
	var identity *yang.Identity = nil
	var mod *Module
	var ymod *yang.Module
	var ok bool
	var base string
	if len(input.Base) > 0 {
		base = input.Base[0].Name
	} else {
		return m, nil
	}
	pre := getPrefix(base)
	name := getName(base)
	if pre != "" {
		mod = getImportedModuleByPrefix(m, pre)
	} else {
		var mname string
		if m.BelongsTo != nil {
			mname = m.BelongsTo.Name
		} else {
			mname = m.Name
		}
		mod, ok = modulesByName[mname]
		if !ok {
			panic("Self module isn't present: " + mname)
		}
	}
	for _, sm := range mod.submodules {
		ymod = sm.module
		for _, e := range ymod.Identity {
			if e.Kind() == "identity" {
				if e.NName() == name {
					identity = e
					break
				}
			}
		}
	}
	return ymod, identity
}

// This function fetches the identity map that is filled in at
// the time of preprocessing modules.
func checkIdentity(m *yang.Module, id *yang.Identity) *yang.Identity {
	var mname string
	if m.BelongsTo != nil {
		mname = m.BelongsTo.Name
	} else {
		mname = m.Name
	}
	if mod, ok := modulesByName[mname]; ok {
		if x, ok := mod.identities[id.Name]; ok {
			return x
		}
	}
	return nil
}

// This function generates all the code needed for each identity
func processIdentity(w io.Writer, submod *SubModule, ymod *yang.Module, n yang.Node) {
	id, ok := n.(*yang.Identity)
	if !ok {
		panic("Not of type Identity")
	}

	// The intent is to see if this identity requires a type defintion
	// Current approach is to generate type definition if any other
	// identity uses this as base. For now, we will always add the
	// absolute base by default. However, the check here needs to happen
	// and is resolved as base is also added to the list
	if id1 := checkIdentity(ymod, id); id1 == id {
		generateTypeDef(w, ymod, id)
	} else {
		if id1 != nil {
			debuglog("processIdentity(): not generating for %s.%s due to %s.%s",
			id.NName(), id.Kind(), id1.NName(), id1.Kind())
		} else {
			debuglog("processIdentity(): not generating for %s.%s due to nil",
			id.NName(), id.Kind())
		}
	}

	// Generate other code related to filling up the maps used in
	// marshal/unmarshal functions
	generateMapEntries(ymod, id)
}

// The map is used to translate the enumeration values generated to strings and
// strings to enumerated values during the marshaling and unmarshaling of the
// the structures. The entries are created for each identity
// Recursively identifies all the base identities and adds code
// for filling up the respective maps
func generateMapEntries(ymod *yang.Module, id *yang.Identity) {
	// Locate the first base and its prefix. This information is used
	// in filling up all the maps as we traverse recursively
	baseymod, id1 := locateBase(ymod, id)
	basemod := getMyModule(baseymod)
	namespace := basemod.namespace
	prefix := basemod.prefix
	for id1 != nil {
		addMapEntry(ymod, id, baseymod, id1, namespace, prefix)
		baseymod, id1 = locateBase(baseymod, id1)
	}
}

// This function adds necessary entries into the maps for a single identity
// statement.
func addMapEntry(m *yang.Module, id *yang.Identity, mbase *yang.Module, base *yang.Identity, namespace string, prefix string) {
	submod := getSubModule(m.Name)
	if submod != nil {
		s := fmt.Sprintf("%s_prefix_map[\"%s\"] = \"%s\"\n", genTN(mbase, base.Name) + "_id", id.Name, prefix)
		submod.initfunc = append(submod.initfunc, s)
		s = fmt.Sprintf("%s_ns_map[\"%s\"] = \"%s\"\n", genTN(mbase, base.Name) + "_id", id.Name, namespace)
		submod.initfunc = append(submod.initfunc, s)
		return
	}
	errorlog("addMapEntry(): Module %s not found for %s.%s", m.Name, id.NName(), id.Kind())
}

// Add the base identity to the module so that it is known that
// code must be generated for this identity
func addBaseIdentity(m *yang.Module, id *yang.Identity) {
	var mod *Module
	var ok bool
	if m.BelongsTo != nil {
		mod, ok = modulesByName[m.BelongsTo.Name]
	} else {
		mod, ok = modulesByName[m.Name]
	}
	if !ok {
		panic("Couldn't locate module" + m.Name)
	}

	mod.addBaseIdentity(id)
}

// The preprocess of identities collects the identities and their respective
// base identities. This will help determine if a data type needs to be
// generated for a given identity
func (m *Module) preprocessIdentities() {
	debuglog("preprocessIdentiites(): processing module %s", m.name)
	for _, sm := range m.submodules {
		for _, i := range sm.module.Identity {
			debuglog("preprocessIdentities(): processing %s.%s", i.NName(), i.Kind())
			if len(i.Base) != 0 {
				// If the identity has a base, locate it
				mod, id := locateBase(sm.module, i)
				if id == nil {
					errorlog("Base couldn't be located %s.%s", i.NName(), i.Kind())
					continue
				}
				// Add the base to the module to be used when
				// code is generated
				addBaseIdentity(mod, id)
			} else {
				m.addBaseIdentity(i)
			}
		}
	}
}
