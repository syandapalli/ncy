package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

// This function returns the type as is in the yang specification
// and does not modify it to suit for coding. This is used when
// resolving the module, etc.
func getType(m *yang.Module, t *yang.Type) string {
	p := t.ParentNode()
	switch t.Name {
	case "leafref":
		ref := getLeafref(t.Path.Name, m, p)
		if ref == nil {
			errorlog("Couldn't locate the leafref for module %s, type %s, path %s", m.Name, t.Name, t.Path.Name)
			return ""
		}
		return ref.Type.Name
	case "identityref":
		return t.IdentityBase.Name
	default:
		return t.Name
	}
}

// This function generates type names for yang keyword "type" used
// to set type to a field. The known field types that use "type"
// are typedef, leaf, leaf-list and deviate. We don't yet support
// deviate and support the rest.
func getTypeName(m *yang.Module, t *yang.Type) string {
	debuglog("getTypeName(): type name for %s.%s in module %s", t.NName(), t.Kind(), m.NName())
	p := t.ParentNode()
	switch t.Name {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "string":
		if t.Range == nil && t.ParentNode().NName() != "union" {
			return t.Name
		} else {
			if p.NName() == "union" {
				id := getIndex(t)
				return genTN(m, fullName(p)) + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
			} else {
				return genTN(m, fullName(p))
			}
		}
	case "decimal8", "decimal16", "decimal32", "decimal64":
		if t.ParentNode().NName() != "union" {
			return "float64"
		} else {
			if p.NName() == "union" {
				id := getIndex(t)
				return genTN(m, fullName(p)) + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
			} else {
				return genTN(m, fullName(p))
			}
		}
	case "leafref":
		ref := getLeafref(t.Path.Name, m, p)
		if ref == nil {
			errorlog("Couldn't locate the leafref for module %s, type %s, path %s", m.Name, t.Name, t.Path.Name)
			return ""
		}
		return getTypeName(getMyYangModule(ref), ref.Type)
	case "boolean":
		if t.ParentNode().NName() == "union" {
			id := getIndex(t)
			return genTN(m, fullName(p)) + "_" + "bool" + "_" + strconv.FormatInt(int64(id), 10)
		}
		return "bool"
	case "enumeration", "union":
		return genTN(m, fullName(p))
	case "binary":
		if p.NName() != "union" {
			return genTN(m, fullName(p))
		} else {
			id := getIndex(t)
			return genTN(m, fullName(p)) + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
		}
	case "identityref":
		return genTN(m, t.IdentityBase.Name + "_id")
	default:
		prefix := getPrefix(t.Name)
		ymod := getMyYangModule(t)
		if mod := getImportedModuleByPrefix(ymod, prefix); mod == nil {
			return ""
		} else {
			return genTN(ymod, t.Name)
		}
	}
}

// This function is responsible for generation of golang type defintions
// for all inclusions of "type". The generation of golang type name
// is handled above.
func processType(w io.Writer, m *yang.Module, n yang.Node) {
	t, ok := n.(*yang.Type)
	if !ok {
		errorlog("processType(): %s.%s is not a Type", n.NName(), n.Kind())
		return
	}
	switch t.Name {
	case "enumeration":
		processEnumType(w, m, t)
	case "leafref":
		processLeafref(w, m, t)
	case "string":
		processStringType(w, m, t)
	case "union":
		processUnionType(w, m, t)
	case "int8", "int16", "int32", "int64":
		processIntType(w, m, t)
	case "uint8", "uint16", "uint32", "uint64":
		processUintType(w, m, t)
	case "decimal8", "decimal16", "decimal32", "decimal64":
		processDecimalType(w, m, t)
	case "boolean":
		processBoolType(w, m, t)
	case "binary":
		processBinaryType(w, m, t)
	case "bits":
		processBitsType(w, m, t)
	case "identityref":
		processIdentityRef(w, m, t)
	default:
		processDefaultType(w, m, t)
	}
}

// This is the case for a non-builtin type being used. As the this new type
// may be referenced elsewhere and may require its own marshal and unmarshal
//functions, we will implement them here
func processDefaultType(w io.Writer, m *yang.Module, t *yang.Type) {
	p := t.ParentNode()
	if p.Kind() != "typedef" {
		debuglog("processDefaultType(): parent node %s.%s isn't a typedef", p.NName(), p.Kind())
		return
	}

	// Basic type definition
	dtn := genTN(m, fullName(p))
	otn := genTN(m, t.Name)
	fmt.Fprintf(w, "type %s %s\n", dtn, otn)

	// Marshal Function
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", dtn)
	fmt.Fprintf(w, "\treturn %s(x).MarshalText(ns)\n", otn)
	fmt.Fprintf(w, "}\n")

	// Unmarshal Function
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", dtn)
	fmt.Fprintf(w, "\treturn ((*%s)(x)).UnmarshalText(ns, b)\n", otn)
	fmt.Fprintf(w, "}\n")
}

func processBoolType(w io.Writer, m *yang.Module, t *yang.Type) {
	// check if we can leave from here as we defined the type.
	// Normally we don't need any marshaling related code if
	// the declaration isn't part of a union where we depend on
	// the type supporting UnmarshalText() and MarshalText()
	// This requires some more thought as the type may be borrowed
	// by a union later.
	p := t.ParentNode()
	if p.Kind() != "typedef" && p.Kind() != "type" {
		return
	}

	// Generate the type name as this is part of a typedef and requires
	// a type definition
	tn := genTN(m, fullName(t.ParentNode()))
	// If this is part of union, to make it unique, we need
	// to add the type name to the full name
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + "bool" + "_" + strconv.FormatInt(int64(id), 10)
	}

	// We now have everything to be able to generate the code
	fmt.Fprintf(w, "type %s bool\n", tn)

	// Generate the marshal code. For this, we need to work with
	// any constraints in the form of range
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	fmt.Fprintf(w, "\treturn []byte(strconv.FormatBool(bool(x))), nil\n")
	fmt.Fprintf(w, "}\n")

	// Generate the unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tv, err := strconv.ParseBool(string(b))\n")
	fmt.Fprintf(w, "\tif err != nil {\n")
	fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\t*x = %s(v)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func processBitsType(w io.Writer, m *yang.Module, t *yang.Type) {
	p := t.ParentNode()
	if p.Kind() != "typedef" && p.Kind() != "type" {
		return
	}

	// Generate the type name as this is part of a typedef and requires
	// a type definition
	tn := genTN(m, fullName(t.ParentNode()))

	// We now have everything to be able to generate the code
	fmt.Fprintf(w, "type %s uint64\n", tn)
}

func processUintType(w io.Writer, m *yang.Module, t *yang.Type) {
	// Check if we need to generate any type definition. We use the
	// golang built-in types where possible. For this type, we need
	// a special type only if the type has additional constraints
	// If not, we can just leave from here
	p := t.ParentNode()
	if t.Range == nil && (p.Kind() != "typedef" && p.Kind() != "type") {
		return
	}

	// var min, max string
	var parts []string
	typestr := t.Name
	sizestr := strings.ReplaceAll(typestr, "uint", "")
	size, err := strconv.ParseUint(sizestr, 10, 8)
	if err != nil {
		panic("Invalid size for uint: " + sizestr)
	}
	tn := genTN(m, fullName(p))
	// If this is part of union, to make it unique, we need
	// to add the type name to the full name
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
	}

	// We now have everything to be able to generate the code
	fmt.Fprintf(w, "type %s uint%s\n", tn, sizestr)

	// Work out the constraints first
	if t.Range != nil {
		rangestr := t.Range.Name
		parts = strings.Split(rangestr, "..")
		if len(parts) != 2 {
			errorlog("processUintType(): Range without two values: %s in %s.%s", t.Range.Name, t.NName(), t.Kind())
			return
		}
	}

	// Generate the marshal code. For this, we need to work with
	// any constraints in the form of range
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	if t.Range != nil && parts[0] != "min" {
		fmt.Fprintf(w, "\tif x < %s {\n", parts[0])
		fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid value %%d for %s\", x)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	if t.Range != nil && parts[1] != "max" {
		fmt.Fprintf(w, "\tif x > %s {\n", parts[1])
		fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid value %%d for %s\", x)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\treturn []byte(strconv.FormatUint(uint64(x), 10)), nil\n")
	fmt.Fprintf(w, "}\n")

	// Generate the unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tv, err := strconv.ParseUint(string(b), 10, %d)\n", size)
	fmt.Fprintf(w, "\tif err != nil {\n")
	fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
	fmt.Fprintf(w, "\t}\n")
	if t.Range != nil && parts[0] != "min" {
		fmt.Fprintf(w, "\tif v < %s {\n", parts[0])
		fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	if t.Range != nil && parts[1] != "max" {
		fmt.Fprintf(w, "\tif v > %s {\n", parts[1])
		fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\t*x = %s(v)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func processIntType(w io.Writer, m *yang.Module, t *yang.Type) {
	// Check if we need to generate any type definition. We use the
	// golang built-in types where possible. For this type, we need
	// a special type only if the type has additional constraints
	// If not, we can just leave from here
	p := t.ParentNode()
	if t.Range == nil && p.Kind() != "typedef" && p.Kind() != "type" {
		return
	}

	// var min, max string
	var parts []string
	typestr := t.Name
	sizestr := strings.ReplaceAll(typestr, "int", "")
	size, err := strconv.ParseUint(sizestr, 10, 8)
	if err != nil {
		panic("Invalid size for int: " + sizestr)
	}
	tn := genTN(m, fullName(p))
	// If this is part of union, to make it unique, we need
	// to add the type name to the full name
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
	}

	// We now have everything to be able to generate the code
	fmt.Fprintf(w, "type %s int%s\n", tn, sizestr)

	// Work out the constraints first
	if t.Range != nil {
		rangestr := t.Range.Name
		parts = strings.Split(rangestr, "..")
		if len(parts) != 2 {
			panic("Range without two values: " + rangestr)
		}
	}

	// Generate the marshal code. For this, we need to work with
	// any constraints in the form of range
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	if t.Range != nil && parts[0] != "min" {
		fmt.Fprintf(w, "\tif x < %s {\n", parts[0])
		fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid value %%d for %s\", x)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	if t.Range != nil && parts[1] != "max" {
		fmt.Fprintf(w, "\tif x > %s {\n", parts[1])
		fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid value %%d for %s\", x)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\treturn []byte(strconv.FormatInt(int64(x), 10)), nil\n")
	fmt.Fprintf(w, "}\n")

	// Generate the unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tv, err := strconv.ParseInt(string(b), 10, %d)\n", size)
	fmt.Fprintf(w, "\tif err != nil {\n")
	fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
	fmt.Fprintf(w, "\t}\n")
	if t.Range != nil && parts[0] != "min" {
		fmt.Fprintf(w, "\tif v < %s {\n", parts[0])
		fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	if t.Range != nil && parts[1] != "max" {
		fmt.Fprintf(w, "\tif v > %s {\n", parts[1])
		fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid value %%s for %s\", string(b))\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\t*x = %s(v)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

// Process decimal type and generate the needed type definitions and marshal
// related code.
func processDecimalType(w io.Writer, m *yang.Module, t *yang.Type) {
	// Check if we need to generate any type definition. We use the
	// golang built-in types where possible. For this type, we need
	// a special type only if the type has additional constraints
	// If not, we can just leave from here
	p := t.ParentNode()
	if p.Kind() != "typedef" && p.Kind() != "type" {
		return
	}

	// var min, max string
	typestr := t.Name
	sizestr := strings.ReplaceAll(typestr, "decimal", "")
	tn := genTN(m, fullName(t.ParentNode()))
	// If this is part of union, to make it unique, we need
	// to add the type name to the full name
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
	}

	// We now have everything to be able to generate the code
	fmt.Fprintf(w, "type %s float%s\n", tn, sizestr)

	// Work out the constraints first
	if t.FractionDigits != nil {
		fmt.Fprintf(w, "/* TODO: support for fraction digits */\n")
	}

	// Marshal code. Currently, the precision in printing is set to -1 (any).
	// This should be improved by using the fraction digits learnt above
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	fmt.Fprintf(w, "\ts := strconv.FormatFloat(float64(x), 'f', -1, %s)\n", sizestr)
	fmt.Fprintf(w, "\treturn []byte(s), nil\n")
	fmt.Fprintf(w, "}\n")

	// Unmarshal code.
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tf, err := strconv.ParseFloat(string(b), %s)\n", sizestr)
	fmt.Fprintf(w, "\tif err != nil {\n")
	fmt.Fprintf(w, "\t\treturn err\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\t*x = %s(f)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func getIndex(n yang.Node) int {
	var id int
	p := n.ParentNode()
	ut, _ := p.(*yang.Type)
	for i, e := range ut.Type {
		if e == n {
			id = i
			break
		}
	}
	return id
}

// This function handles all types that are "string"
func processStringType(w io.Writer, m *yang.Module, t *yang.Type) {
	// Check if we need to generate any type definition. We use the
	// golang built-in types where possible. For this type, we need
	// a special type only if the type has additional constraints
	// If not, we can just leave from here
	p := t.ParentNode()
	if t.Length == nil && p.Kind() != "typedef" && p.Kind() != "type" {
		return
	}

	// variables to manage constraints
	var len_constraint bool
	var min, max string
	var pattern_constraint bool
	var pattern string

	// First generate the type definition
	tn := genTN(m, fullName(t.ParentNode()))
	// If this is part of union, to make it unique, we need
	// to add the type name to the full name
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
	}

	fmt.Fprintf(w, "type %s string\n", tn)

	// Handle constraints. The strings may need to match some
	// pattern or may be constrained to length
	if t.Length != nil {
		len_constraint = true
		parts := strings.Split(t.Length.Name, "..")
		if len(parts) == 1 {
			min = parts[0]
			max = min
		} else if len(parts) == 2 {
			min = parts[0]
			max = parts[1]
		}
	}
	if len(t.Pattern) > 0 {
		pattern_constraint = true
		pattern = t.Pattern[0].Name
		pattern = strings.ReplaceAll(pattern, "\\", "\\\\")
	}

	// Generate regular expression initialization
	if pattern_constraint {
		fmt.Fprintf(w, "var %s_re = regexp.MustCompile(\"%s\")\n", tn, pattern)
	}

	// Generate MarshalText() that takes into consideration the constraints
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	if len_constraint {
		if min != "" && min != "min" {
			fmt.Fprintf(w, "\tif len(x) < %s {\n", min)
			fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid length %%d for %s\", len(x))\n", tn)
			fmt.Fprintf(w, "\t}\n")
		}
		if max != "" && max != "max" {
			fmt.Fprintf(w, "\tif len(x) > %s {\n", max)
			fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid length %%d for %s\", len(x))\n", tn)
			fmt.Fprintf(w, "\t}\n")
		}
	}
	if pattern_constraint {
		fmt.Fprintf(w, "\tif !%s_re.MatchString(string(x)) {\n", tn)
		fmt.Fprintf(w, "\t\treturn nil, fmt.Errorf(\"Invalid %s: %%s\", x)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\treturn []byte(x), nil\n")
	fmt.Fprintf(w, "}\n")

	// Generate UnmarshalText() that verifies the constraints
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\ts := string(b)\n")
	if len_constraint {
		if min != "" && min != "min" {
			fmt.Fprintf(w, "\tif len(s) < %s {\n", min)
			fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid length %%d for %s\", len(s))\n", tn)
			fmt.Fprintf(w, "\t}\n")
		}
		if max != "" && max != "max" {
			fmt.Fprintf(w, "\tif len(s) > %s {\n", max)
			fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid length %%d for %s\", len(s))\n", tn)
			fmt.Fprintf(w, "\t}\n")
		}
	}
	if pattern_constraint {
		fmt.Fprintf(w, "\tif !%s_re.MatchString(s) {\n", tn)
		fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"Invalid %s: %%s\", s)\n", tn)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\t*x = %s(s)\n", tn)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

// This function generates code for yang union. Union can include
// any one of the types that are part of the union. The decoding
// happens for each type and the first successful type is assumed
// to be the encoded type.
func processUnionType(w io.Writer, m *yang.Module, t *yang.Type) {
	// First generate the fields of the union
	tn := genTN(m, fullName(t.ParentNode()))
	fmt.Fprintf(w, "type %s struct {\n", tn)
	for id, it := range t.Type {
		fmt.Fprintf(w, "\t%s_%d_Prsnt bool `xml:\",presfield\"`\n", genFN(it.Name), id)
		fmt.Fprintf(w, "\t%s_%d %s `xml:\"-\"`\n", genFN(it.Name), id, getTypeName(m, it))
	}
	fmt.Fprintf(w, "}\n")

	// Generate marshal code
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	for id, it := range t.Type {
		fn := genFN(it.Name)
		fmt.Fprintf(w, "\tif x.%s_%d_Prsnt {\n", fn, id)
		fmt.Fprintf(w, "\t\treturn x.%s_%d.MarshalText(ns)\n", fn, id)
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\treturn nil, fmt.Errorf(\"Invalid %s\")\n", tn)
	fmt.Fprintf(w, "}\n")

	// Generate unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	for id, it := range t.Type {
		fn := genFN(it.Name)
		fmt.Fprintf(w, "\tif err := (&x.%s_%d).UnmarshalText(ns, b); err == nil {\n", fn, id)
		fmt.Fprintf(w, "\t\tx.%s_%d_Prsnt = true\n", fn, id)
		fmt.Fprintf(w, "\t\treturn nil\n")
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\treturn fmt.Errorf(\"Invalid %s: %%s\", string(b))\n", tn)
	fmt.Fprintf(w, "}\n")

	for _, it := range t.Type {
		fmt.Fprintln(w, "/* Generating type for", it.Name, "-parent", t.Kind(), "*/")
		processType(w, m, it)
	}
}

// This function generates golang code for yang enumeration. Yang
// enumeration can either be included in a typedef or inside
// a grouping/container/list without explicit type name. For
// such inclusions, the type name is implicitly derived from
// the place of inclusion
func processEnumType(w io.Writer, m *yang.Module, t *yang.Type) {
	// Sanity check to see if we are OK
	if t.Name != "enumeration" {
		errorlog("processEnumType():%s.%s isn't an enumeration", t.NName(), t.Kind())
		return
	}

	// Generate the type statement for the enum. All enums
	// translate to int in our implementation
	tname := genTN(m, fullName(t.ParentNode()))
	fmt.Fprintf(w, "type %s int\n", tname)

	// Generate the constants for the enums
	fmt.Fprintf(w, "const (\n")
	for i, en := range t.Enum {
		fname := genFN(en.Name)
		if i == 0 {
			fmt.Fprintf(w, "\t%s_%s %s = iota\n", tname, fname, tname)
		} else {
			fmt.Fprintf(w, "\t%s_%s\n", tname, fname)
		}
	}
	fmt.Fprintf(w, ")\n")

	// Generate mapping from the golang enumeration to yang enumerations
	fmt.Fprintf(w, "var %s_to_string = map[%s]string {\n", tname, tname)
	for _, en := range t.Enum {
		fname := genFN(en.Name)
		fmt.Fprintf(w, "\t%s_%s: \"%s\",\n", tname, fname, en.Name)
	}
	fmt.Fprintf(w, "}\n")

	// Generate mapping from the yang enumerations to the golan enumeration
	fmt.Fprintf(w, "var string_to_%s = map[string]%s {\n", tname, tname)
	for _, en := range t.Enum {
		fname := genFN(en.Name)
		fmt.Fprintf(w, "\t\"%s\": %s_%s,\n", en.Name, tname, fname)
	}
	fmt.Fprintf(w, "}\n")

	// Generate Marshal code
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tname)
	fmt.Fprintf(w, "\tif s, ok := %s_to_string[x]; ok {\n", tname)
	fmt.Fprintf(w, "\t\treturn []byte(s), nil\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\treturn nil, fmt.Errorf(\"Invalid value for %s\")\n", tname)
	fmt.Fprintf(w, "}\n")

	// Generate Marshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tname)
	fmt.Fprintf(w, "\tif v, ok := string_to_%s[string(b)]; ok {\n", tname)
	fmt.Fprintf(w, "\t\t*x = v\n")
	fmt.Fprintf(w, "\t\treturn nil\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\treturn fmt.Errorf(\"Invalid value for %s\")\n", tname)
	fmt.Fprintf(w, "}\n")
}

// Process leafref where a path is used to parse the tree to obtain the type
// to be used
func processLeafref(w io.Writer, m *yang.Module, t *yang.Type) {
	p := t.ParentNode()
	if p.Kind() != "leaf" && p.Kind() != "typedef" {
		errorlog("processLeafref(): parent node %s.%s is not a valid kind", p.NName(), p.Kind())
		return
	}
	path := t.Path.Name
	l := getLeafref(path, m, p)
	if l == nil {
		errorlog("processLeafref(): leaf for referrence %s is not found", path)
		return
	}
	fmt.Fprintf(w, "type %s %s\n", genTN(m, fullName(p)), getTypeName(m, l.Type))
}

// Process leafref where a path is used to parse the tree to obtain the type
// to be used
func processIdentityRef(w io.Writer, m *yang.Module, t *yang.Type) {
	// generate the type definition
	p := t.ParentNode()
	if p.Kind() != "typedef" {
		debuglog("processIdentityRef(): %s.%s is not a typedef", p.NName(), p.Kind())
		return
	}
	otn := getTypeName(m, t)
	itn := genTN(m, fullName(p))
	fmt.Fprintf(w, "type %s %s\n", itn, otn)
	// Generate the marshal code
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", itn)
	fmt.Fprintf(w, "\treturn %s(x).MarshalText(ns)\n", otn)
	fmt.Fprintf(w, "}\n")

	// Generate the unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", itn)
	fmt.Fprintf(w, "\treturn ((*%s)(x)).UnmarshalText(ns, b)\n", otn)
	fmt.Fprintf(w, "}\n")
}

// note:
// 	"binary" is a special case till we find a better way of
// 	handling it. Currently just one instance exists for Ieeefloat
// 	and so we are going to hanlde it explicitly
func processBinaryType(w io.Writer, m *yang.Module, t *yang.Type) {
	p := t.ParentNode()
	//if p.Kind() != "typedef" && p.Kind() != "type" {
	//	return
	//}

	// Generate the type name as this is part of a typedef and requires
	// a type definition
	if p.NName() == "ieeefloat32" {
		handleIeeefloat32(w, m, t)
		return
	}

	// Binary type is an array of bytes which should be encoded in
	// base64 encoding.
	// First generate the type definition
	tn := genTN(m, fullName(p))
	if p.Kind() == "type" && p.NName() == "union" {
		id := getIndex(t)
		tn = tn + "_" + t.Name + "_" + strconv.FormatInt(int64(id), 10)
	}
	fmt.Fprintf(w, "type %s []byte\n", tn)

	//Generate Marshal code
	fmt.Fprintf(w, "func (x %s)MarshalText(ns string) ([]byte, error) {\n", tn)
	fmt.Fprintf(w, "\ts := base64.StdEncoding.EncodeToString(x)\n")
	fmt.Fprintf(w, "\treturn []byte(s), nil\n")
	fmt.Fprintf(w, "}\n")

	//Generate Unmarshal code
	fmt.Fprintf(w, "func (x *%s)UnmarshalText(ns string, b []byte) error {\n", tn)
	fmt.Fprintf(w, "\tv, err := base64.StdEncoding.DecodeString(string(b))\n")
	fmt.Fprintf(w, "\t*x = %s(v)\n", tn)
	fmt.Fprintf(w, "\treturn err\n")
	fmt.Fprintf(w, "}\n")
}

func handleIeeefloat32(w io.Writer, m *yang.Module, t *yang.Type) {
	fmt.Fprintf(w, "type Oc_types_ieeefloat32 float32\n")
	fmt.Fprintf(w, "func (x Oc_types_ieeefloat32) MarshalText(ns string) ([]byte, error) {\n")
	fmt.Fprintf(w, "        ix := math.Float32bits(float32(x))\n")
	fmt.Fprintf(w, "        str := strconv.FormatUint(uint64(ix),2)\n")
	fmt.Fprintf(w, "        zeros := 32 - len(str)\n")
	fmt.Fprintf(w, "        i := 0\n")
	fmt.Fprintf(w, "        for (i < zeros) {\n")
	fmt.Fprintf(w, "                str = \"0\" + str\n")
	fmt.Fprintf(w, "                i++\n")
	fmt.Fprintf(w, "        }\n")
	fmt.Fprintf(w, "        return []byte(str), nil\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "func (x *Oc_types_ieeefloat32) UnmarshalText(ns string, b []byte) error {\n")
	fmt.Fprintf(w, "        y, err := strconv.ParseUint(string(b), 2, 32)\n")
	fmt.Fprintf(w, "        if err != nil {\n")
	fmt.Fprintf(w, "                return fmt.Errorf(\"Invalid Oc_types_ieeefloat32\")\n")
	fmt.Fprintf(w, "        }\n")
	fmt.Fprintf(w, "        *x = Oc_types_ieeefloat32(math.Float32frombits(uint32(y)))\n")
	fmt.Fprintf(w, "        return nil\n")
	fmt.Fprintf(w, "}\n")
}
