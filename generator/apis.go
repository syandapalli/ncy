package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

type Elements struct {
	Elements []Element `json:"elements"`
}
type Element struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

type SetStructs struct {
	Structs []SetStructure `json:"set-structs"`
}
type GetStructs struct {
	Structs []GetStructure `json:"get-structs"`
}
type SetBulkStructs struct {
	Structs []BulkStructure `json:"set-bulk"`
}
type GetBulkStructs struct {
	Structs []BulkStructure `json:"get-bulk"`
}
type SetStructure struct {
	Name    string `json:"name"`
	Members string `json:"members"`
}
type GetStructure struct {
	Name    string `json:"name"`
	Filter  string `json:"filter"`
	Members string `json:"members"`
}
type BulkStructure struct {
	Name    string `json:"name"`
	Structs string `json:"structs"`
}

type FileInfo struct {
	Name           string
	Members        Elements
	SetStructs     SetStructs
	GetStructs     GetStructs
	SetBulkStructs SetBulkStructs
	GetBulkStructs GetBulkStructs
}

type NodeInfo struct {
	node   *yang.Node
	mod    *yang.Module
	parent interface{}
}

type NodeData struct {
	node       *yang.Node
	mod        *yang.Module
	varType    string
	varName    string
	kind       string
	structName string
	element    *Element
	parent     interface{}
	children   []interface{}
}

var elementById = map[string]map[int]Element{}
var setStructByName = map[string]*SetStructure{}
var getStructByName = map[string]*GetStructure{}

func addToSetStructMap(name string, s *SetStructure) error {
	if _, ok := setStructByName[name]; ok {
		log.Printf("Struct with same name present: %s", name)
		return fmt.Errorf("struct with same name present: %s", name)
	} else {
		sNew := *s
		setStructByName[name] = &sNew
		return nil
	}
}
func getFromSetStructMap(name string) *SetStructure {
	if s, ok := setStructByName[name]; ok {
		return s
	} else {
		log.Printf("Struct with name %s does not present", name)
		return nil
	}
}
func addToGetStructMap(name string, s *GetStructure) error {
	if _, ok := getStructByName[name]; ok {
		log.Printf("Struct with same name present: %s", name)
		return fmt.Errorf("struct with same name present: %s", name)
	} else {
		sNew := *s
		getStructByName[name] = &sNew
		return nil
	}
}
func getFromGetStructMap(name string) *GetStructure {
	if s, ok := getStructByName[name]; ok {
		return s
	} else {
		log.Printf("Struct with name %s does not present", name)
		return nil
	}
}

func processStructsAndApis(jsonDir string, genDir string) {
	files := readDir(jsonDir, ".json")
	fmt.Println("Number files = ", len(files))

	var fileInfos []*FileInfo
	// Decode the input files and store it in cache
	for _, file := range files {
		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic("Cannot open file: " + err.Error())
		}
		var byteValue []byte
		byteValue, err = ioutil.ReadAll(jsonFile)
		if err != nil {
			panic("Cannot open file: " + err.Error())
		}

		// get the file name:
		_, file := path.Split(file)
		outpath := genDir + "/api/" + file
		outpath = strings.ReplaceAll(outpath, ".json", "")
		ensureDirectory(outpath)
		fmt.Println("Processing file", file, "...")

		var fileInfo FileInfo
		fileInfo.Name = outpath
		if err := json.Unmarshal(byteValue, &fileInfo.Members); err != nil {
			panic("Cannot unmarshal file: " + err.Error())
		}
		if err := json.Unmarshal(byteValue, &fileInfo.SetStructs); err != nil {
			panic("Cannot unmarshal file: " + err.Error())
		}
		if err := json.Unmarshal(byteValue, &fileInfo.GetStructs); err != nil {
			panic("Cannot unmarshal file: " + err.Error())
		}
		if err := json.Unmarshal(byteValue, &fileInfo.SetBulkStructs); err != nil {
			panic("Cannot unmarshal file: " + err.Error())
		}
		fileInfos = append(fileInfos, &fileInfo)
	}

	// Update the map for elements for faster search later
	for _, fileInfo := range fileInfos {
		for _, elem := range fileInfo.Members.Elements {
			if _, ok := elementById[fileInfo.Name]; !ok {
				elementById[fileInfo.Name] = map[int]Element{}
			}
			if _, ok := elementById[fileInfo.Name][elem.Id]; ok {
				panic("Element with same id")
			}
			elementById[fileInfo.Name][elem.Id] = elem
		}
	}
	for _, fileInfo := range fileInfos {
		generateStructsAndApis(fileInfo, genDir)
	}
}

func generateStructsAndApis(fileInfo *FileInfo, yangGoDir string) {
	// Open a file for writing the structs
	ws, err := os.OpenFile(fileInfo.Name+"_structs.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err.Error())
	}
	// Open a file for writing the apis
	wa, err1 := os.OpenFile(fileInfo.Name+"_apis.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err1 != nil {
		panic(err1.Error())
	}
	_, file := path.Split(fileInfo.Name)
	structFileHeader(ws)
	apiFileHeader(wa, file, yangGoDir)
	mergeModules()

	for _, s := range fileInfo.SetStructs.Structs {
		processSetStruct(ws, fileInfo.Name, &s)
		processSetApi(wa, fileInfo.Name, &s)
	}

	for _, s := range fileInfo.GetStructs.Structs {
		processGetStruct(ws, fileInfo.Name, &s)
		processGetApi(wa, fileInfo.Name, &s)
	}
	for _, s := range fileInfo.SetBulkStructs.Structs {
		processBulkSetStruct(ws, fileInfo.Name, &s)
		processBulkSetApi(wa, fileInfo.Name, &s)
	}
	ws.Close()
	wa.Close()
}

func structFileHeader(w io.Writer) {
	// Generic header of the file with imports and package
	// In future, we should take package name as an attribute
	fmt.Fprintf(w, "// Code generated by ncgenerate\n")
	fmt.Fprintf(w, "package api\n")
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, ")\n\n")
}

func apiFileHeader(w io.Writer, dummy, yangGoDir string) {
	yangGoAbsDir, err := filepath.Abs(yangGoDir)
	if err != nil {
		log.Fatalf("yangGoDir is not valid %s", yangGoDir)
	}
	parts := strings.Split(yangGoAbsDir, "src/")
	yangGoPath := parts[0]
	if len(parts) == 2 {
		yangGoPath = parts[1]
	}
	yangGoPath += "/yang-go"
	log.Println("Yang-go path:", yangGoPath)

	// Generic header of the file with imports and package
	// In future, we should take package name as an attribute
	fmt.Fprintf(w, "// Code generated by ncgenerate\n")
	fmt.Fprintf(w, "package api\n")
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, "\t\"fmt\"\n")
	fmt.Fprintf(w, "\t\"log\"\n")
	fmt.Fprintf(w, "\toc \"%s\"\n", yangGoPath)
	fmt.Fprintf(w, "\t\"toradapter/lib/netconf\"\n")
	fmt.Fprintf(w, "\tnc \"toradapter/lib/encoding/nc\"\n")
	fmt.Fprintf(w, "\tproto \"toradapter/lib/netconf/protocol\"\n")
	fmt.Fprintf(w, ")\n")
	// Generate the dummy usage for all the common import packages so that
	// we don't have to carefully identify which of them to be included

	dummy = strings.ReplaceAll(dummy, "-", "_")
	fmt.Fprintf(w, "\n//-----------------------------------------------------\n")
	fmt.Fprintf(w, "//Dummy code to avoid careful insertion of imports\n")
	fmt.Fprintf(w, "var %s_err = fmt.Errorf(\"dummy\")\n", dummy)
	fmt.Fprintf(w, "//-----------------------------------------------------\n\n")

}

//--------------------- Functions related to set operations ---------------------------
func processSetStruct(ws io.Writer, fileName string, s *SetStructure) {
	fmt.Fprintf(ws, "type Set%s struct {\n", s.Name)
	addStructMember(ws, fileName, s.Name, strings.Split(s.Members, " "), true)
	fmt.Fprintf(ws, "}\n")
	// Add struct to map.
	addToSetStructMap(s.Name, s)
}

func addStructMember(ws io.Writer, fileName, sname string, members []string, isSetStruct bool) {
	var elementByName map[string]Element = map[string]Element{}
	for _, m := range members {
		id, err := strconv.Atoi(m)
		if err != nil {
			log.Fatalf("Invalid id in struct %s %s", m, sname)
		}
		e, ok := elementById[fileName][id]
		if !ok {
			log.Fatalf("Invalid id in struct %s %s. Not present in elements", m, sname)
		}
		if val, ok := elementByName[e.Name]; ok {
			log.Printf("Element with same name: '%s' in struct '%s'", e.Name, sname)
			// If element with same name found then check the type. They should match
			if e.Type != val.Type {
				log.Fatalf("Element with same name but different types found '%s'", e.Name)
			}
		} else if e.Value == "" {
			elementByName[e.Name] = e
			if isSetStruct {
				fmt.Fprintf(ws, "\t%s_Prsnt bool\n", e.Name)
			}
			fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
		}
	}
}

func processSetApi(w io.Writer, fileName string, s *SetStructure) {
	fmt.Fprintf(w, "\nfunc ProcessSet%s(sID int, v *Set%s) error {\n", s.Name, s.Name)
	fmt.Fprintf(w, "\n\tvar buffer, b []byte\n")
	fmt.Fprintf(w, "\tvar err error\n")
	processSetApiMembers(w, fileName, s, "err")
	addSetRPCCall(w)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func processSetApiMembers(w io.Writer, fileName string, s *SetStructure, returnVal string) {
	initNodesTree()
	members := strings.Split(s.Members, " ")
	for _, m := range members {
		id, _ := strconv.Atoi(m)
		e := elementById[fileName][id]
		// fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
		updateNodesTreeFromPath(e.Path, s.Name, &e)
	}

	//fmt.Println("No of root nodes", len(RootNodes))
	for _, nodeInfo := range RootNodes {
		fmt.Fprintf(w, "\t{\n")
		str := ""
		processNodeForSetBulk(w, nodeInfo, &str)

		fmt.Fprintf(w, "\t\tb, err = nc.MarshalIndent(&%s, \"\", \"  \")\n", nodeInfo.varName)
		fmt.Fprintf(w, "\t\tif err != nil {\n")
		fmt.Fprintf(w, "\t\t\tlog.Printf(\"Failed to edit config: %%s\", err.Error())\n")
		fmt.Fprintf(w, "\t\t\treturn %s\n", returnVal)
		fmt.Fprintf(w, "\t\t}\n")
		fmt.Fprintf(w, "\t\tbuffer = append(buffer, b...)\n")
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "\n")
}

func addSetRPCCall(w io.Writer) {
	rpcString := `
	if len(buffer) > 0 {
		rpc := proto.NewRPCMessage()
		rpc.Edit_Config_Prsnt = true
		rpc.Edit_Config.Target.Candidate_Prsnt = true
		rpc.Edit_Config.Target.Candidate = true
		rpc.Edit_Config.Config.XmlString = string(buffer)
		reply, err1 := netconf.ExecuteRpc(sID, rpc)
		if err1 != nil {
			log.Println("RPC Error:", err1)
			return err1
		}
		if reply.Error_Prsnt {
			log.Println("RPC Reply error: ", reply.Error)
			return fmt.Errorf("RPC Reply error: %s", reply.Error.Error())
		}

		rpc = proto.NewRPCMessage()
		rpc.Commit_Prsnt = true
		rpc.Commit = true
		reply, err1 = netconf.ExecuteRpc(sID, rpc)
		if err1 != nil {
			log.Println("RPC Error:", err1)
			return err1
		}
		if reply.Error_Prsnt {
			log.Println("RPC Reply error: ", reply.Error)
			return fmt.Errorf("RPC Reply error: %s", reply.Error.Error())
		}
	}`
	fmt.Fprintf(w, "%s\n", rpcString)
}

//--------------------- Functions related to Get operations ---------------------------
func processGetStruct(ws io.Writer, fileName string, s *GetStructure) {
	fmt.Fprintf(ws, "type Filter%s struct {\n", s.Name)
	addStructMember(ws, fileName, s.Name, strings.Split(s.Filter, " "), true)
	fmt.Fprintf(ws, "}\n\n")
	fmt.Fprintf(ws, "type Get%s struct {\n", s.Name)
	addStructMember(ws, fileName, s.Name, strings.Split(s.Members, " "), true)
	fmt.Fprintf(ws, "}\n")
	addToGetStructMap(s.Name, s)
}

func processGetApi(w io.Writer, fileName string, s *GetStructure) {
	fmt.Fprintf(w, "\nfunc ProcessGet%s(sID int, v *Filter%s) *Get%s {\n", s.Name, s.Name, s.Name)
	processGetRequestApi(w, fileName, s)
	processGetReplyApi(w, fileName, s)
	fmt.Fprintf(w, "\treturn &g\n")
	fmt.Fprintf(w, "}\n")
}

func processGetRequestApi(w io.Writer, fileName string, s *GetStructure) {
	ss := &SetStructure{s.Name, s.Filter}
	fmt.Fprintf(w, "\n\tvar buffer, b []byte\n")
	fmt.Fprintf(w, "\tvar err error\n")
	fmt.Fprintf(w, "\tvar reply *proto.RPCReply\n")
	processSetApiMembers(w, fileName, ss, "nil")
	addGetRPCCall(w)
}

func addGetRPCCall(w io.Writer) {
	rpcString := `
	if len(buffer) > 0 {
		rpc := proto.NewRPCMessage()
		rpc.Get_Prsnt = true
		rpc.Get.Filter_Prsnt = true
		rpc.Get.Filter.Type = "subtree"
		rpc.Get.Filter.Transparent = string(buffer)
		reply, err = netconf.ExecuteRpc(sID, rpc)
		if err != nil {
			log.Println("RPC Error:", err)
			return nil
		}
		if reply.Error_Prsnt {
			log.Println("RPC Reply error: ", reply.Error)
			return nil
		}
	} else {
		return nil
	}
`
	fmt.Fprintf(w, "%s\n", rpcString)
}

func processGetReplyApi(w io.Writer, fileName string, s *GetStructure) {
	// Create a global struct to unmarshal.
	fmt.Fprintf(w, "\n\ttype Data struct {\n")
	for _, nodeInfo := range RootNodes {
		nodeInfo.processNodeForUnmarshalStruct(w)
		// fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
	}
	fmt.Fprintf(w, "\t}\n\n")

	// Unmarshal code generation
	fmt.Fprintf(w, "\tvar data Data\n")
	fmt.Fprintf(w, "\tif reply.Data_Prsnt {\n")
	fmt.Fprintf(w, "\t\td := []byte(\"<data>\"+reply.Data.Innerxml+\"</data>\")\n")
	fmt.Fprintf(w, "\t\tvar err error\n")
	fmt.Fprintf(w, "\t\terr = nc.Unmarshal(d, &data, true)\n")
	fmt.Fprintf(w, "\t\tif err != nil {\n")
	fmt.Fprintf(w, "\t\t\tlog.Println(\"Unmarshal Error:\", err)\n")
	fmt.Fprintf(w, "\t\t\treturn nil\n")
	fmt.Fprintf(w, "\t\t}\n")
	//fmt.Fprintf(w, "\t\tfmt.Printf(\"%%+v\\n\", %s)\n", rootVar)
	fmt.Fprintf(w, "\t} else {\n")
	fmt.Fprintf(w, "\t\treturn nil\n")
	fmt.Fprintf(w, "\t}\n\n")

	fmt.Fprintf(w, "\tvar g Get%s\n", s.Name)
	initNodesTree()
	members := strings.Split(s.Members, " ")
	for _, m := range members {
		id, _ := strconv.Atoi(m)
		e := elementById[fileName][id]
		// fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
		updateNodesTreeFromPath(e.Path, s.Name, &e)
	}

	for _, nodeInfo := range RootNodes {
		processNodeForGet(w, nodeInfo)
	}
}

func (n *NodeData) processNodeForUnmarshalStruct(w io.Writer) {
	switch n.kind {
	case "container":
		fmt.Fprintf(w, "\t\t%s oc.%s\n", n.varName, n.varType)
	default:
		panic("Never reach here")
	}
}

func processNodeForGet(w io.Writer, n *NodeData) {
	switch n.kind {
	case "leaf":
		if len(n.children) > 0 {
			// There may be union present next. Return from here
			break
		}
		parentVarName := (n.parent.(*NodeData)).varName
		preCheck := ""
		if strings.Contains(parentVarName, "[0]") {
			preCheck = "len(data." + strings.Split(parentVarName, "[0]")[0] + ") == 1 && "
		}
		fmt.Fprintf(w, "\tif %sdata.%s_Prsnt {\n", preCheck, n.varName)
		fmt.Fprintf(w, "\t\tg.%s_Prsnt = true\n", n.element.Name)
		fmt.Fprintf(w, "\t\tg.%s = %s(data.%s)\n", n.element.Name, n.element.Type, n.varName)
		fmt.Fprintf(w, "\t}\n")
	case "union":
		// before union one leaf will be there.
		p := n.parent.(*NodeData)
		parentVarName := (p.parent.(*NodeData)).varName
		preCheck := ""
		if strings.Contains(parentVarName, "[0]") {
			preCheck = "len(data." + strings.Split(parentVarName, "[0]")[0] + ") == 1 && "
		}
		fmt.Fprintf(w, "\tif %sdata.%s_Prsnt {\n", preCheck, n.varName)
		fmt.Fprintf(w, "\t\tg.%s_Prsnt = true\n", n.element.Name)
		fmt.Fprintf(w, "\t\tg.%s = %s(data.%s)\n", n.element.Name, n.element.Type, n.varName)
		fmt.Fprintf(w, "\t}\n")
	}
	for _, child := range n.children {
		c := child.(*NodeData)
		processNodeForGet(w, c)
	}
}

//--------------------- Functions related to Set-bulk operations ---------------------------
func processBulkStruct(ws io.Writer, fileName string, bs *BulkStructure, structPrefix string) {
	fmt.Fprintf(ws, "type %sBulk%s struct {\n", structPrefix, bs.Name)
	var structByName map[string]struct{} = map[string]struct{}{}
	members := strings.Split(bs.Structs, " ")
	for _, m := range members {
		if s := getFromSetStructMap(m); s == nil {
			log.Fatalf("Invalid struct name %s", m)
		}
		if _, ok := structByName[m]; ok {
			log.Fatalf("Duplicate struct name %s", m)
		}
		structByName[m] = struct{}{}
		fmt.Fprintf(ws, "\t%s []%s%s\n", m, structPrefix, m)
	}
	fmt.Fprintf(ws, "}\n")
}

func processBulkSetStruct(ws io.Writer, fileName string, bs *BulkStructure) {
	processBulkStruct(ws, fileName, bs, "Set")
}

func processBulkSetApi(w io.Writer, fileName string, bs *BulkStructure) {
	fmt.Fprintf(w, "\nfunc ProcessSetBulk%s(sID int, v *SetBulk%s) error {\n", bs.Name, bs.Name)
	fmt.Fprintf(w, "\n\tvar buffer, b []byte\n")
	fmt.Fprintf(w, "\tvar err error\n")
	processBulkMembers(w, fileName, bs, true)
	addSetRPCCall(w)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func processBulkMembers(w io.Writer, fileName string, bs *BulkStructure, isSetOp bool) {
	members := strings.Split(bs.Structs, " ")
	for _, m := range members {
		mems := ""
		if isSetOp {
			mems = getFromSetStructMap(m).Members
		} else {
			mems = getFromGetStructMap(m).Filter
		}
		membersInside := strings.Split(mems, " ")
		initNodesTree()
		for _, mIn := range membersInside {
			id, _ := strconv.Atoi(mIn)
			e := elementById[fileName][id]
			// fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
			updateNodesTreeFromPath(e.Path, m, &e)
		}

		for _, nodeInfo := range RootNodes {
			fmt.Fprintf(w, "\t{\n")

			processNodeForSetBulk(w, nodeInfo, nil)

			fmt.Fprintf(w, "\t\tif len(v.%s) > 0 {\n", nodeInfo.structName)
			fmt.Fprintf(w, "\t\t\tb, err = nc.MarshalIndent(&%s, \"\", \"  \")\n", nodeInfo.varName)
			fmt.Fprintf(w, "\t\t\tif err != nil {\n")
			fmt.Fprintf(w, "\t\t\t\tlog.Printf(\"Failed to edit config: %%s\", err.Error())\n")
			fmt.Fprintf(w, "\t\t\t\treturn err\n")
			fmt.Fprintf(w, "\t\t\t}\n")
			fmt.Fprintf(w, "\t\t\tbuffer = append(buffer, b...)\n")
			fmt.Fprintf(w, "\t\t}\n")
			fmt.Fprintf(w, "\t}\n")
		}
		fmt.Fprintf(w, "\n\n")
	}
}

func processNodeForSetBulk(w io.Writer, n *NodeData, loopVar *string) {
	addedLoop := false
	userStruct := ""
	tab := ""
	if loopVar != nil && *loopVar != "" {
		userStruct = n.structName + "[i]."
		tab = "\t"
		n.varName = strings.ReplaceAll(n.varName, (*loopVar)+"[0]", (*loopVar)+"[i]")
	}
	switch n.kind {
	case "container":
		if n.parent == nil {
			fmt.Fprintf(w, "\t\tvar %s oc.%s\n", n.varName, n.varType)
		} else {
			fmt.Fprintf(w, "%s\t\t%s_Prsnt = true\n", tab, n.varName)
		}
	case "list", "leaf-list":
		// start a loop if required.
		if loopVar == nil && checkIfLoopRequired(n) {
			fmt.Fprintf(w, "\t\tfor i, _ := range v.%s {\n", n.structName)
			loopVar = new(string)
			*loopVar = strings.TrimSuffix(n.varName, "[0]")
			addedLoop = true
			tab = "\t"
		}
		v := strings.TrimSuffix(n.varName, "[0]")
		vars := strings.Split(v, ".")
		l := len(vars)
		varName := vars[l-1]
		fmt.Fprintf(w, "\t\t%svar %s oc.%s\n", tab, varName, n.varType)
		fmt.Fprintf(w, "\t\t%s%s = append(%s, %s)\n", tab, v, v, varName)
	case "leaf":
		if len(n.children) > 0 {
			// There may be union present next. Return from here
			fmt.Fprintf(w, "%s\t\t%s_Prsnt = true\n", tab, n.varName)
			break
		}
		if n.element.Value != "" {
			fmt.Fprintf(w, "%s\t\t%s_Prsnt = true\n", tab, n.varName)
			if n.element.Type == "string" {
				fmt.Fprintf(w, "%s\t\t%s = %s(\"%s\")\n", tab, n.varName,
					addPrefixToDatatypeIfRequired(n.varType, "oc"), n.element.Value)
			} else {
				fmt.Fprintf(w, "%s\t\t%s = %s(%s)\n", tab, n.varName,
					addPrefixToDatatypeIfRequired(n.varType, "oc"), n.element.Value)
			}
		} else {
			fmt.Fprintf(w, "%s\t\tif v.%s%s_Prsnt {\n", tab, userStruct, n.element.Name)
			fmt.Fprintf(w, "%s\t\t\t%s_Prsnt = true\n", tab, n.varName)
			fmt.Fprintf(w, "%s\t\t\t%s = %s(v.%s%s)\n", tab, n.varName, addPrefixToDatatypeIfRequired(n.varType, "oc"),
				userStruct, n.element.Name)
			fmt.Fprintf(w, "%s\t\t}\n", tab)
		}
	case "union":
		if n.element.Value != "" {
			fmt.Fprintf(w, "%s\t\t%s_Prsnt = true\n", tab, n.varName)
			if n.element.Type == "string" {
				fmt.Fprintf(w, "%s\t\t%s = %s(\"%s\")\n", tab, n.varName,
					addPrefixToDatatypeIfRequired(n.varType, "oc"), n.element.Value)
			} else {
				fmt.Fprintf(w, "%s\t\t%s = %s(%s)\n", tab, n.varName,
					addPrefixToDatatypeIfRequired(n.varType, "oc"), n.element.Value)
			}
		} else {
			fmt.Fprintf(w, "%s\t\tif v.%s%s_Prsnt {\n", tab, userStruct, n.element.Name)
			fmt.Fprintf(w, "%s\t\t\t%s_Prsnt = true\n", tab, n.varName)
			fmt.Fprintf(w, "%s\t\t\t%s = %s(v.%s%s)\n", tab, n.varName, addPrefixToDatatypeIfRequired(n.varType, "oc"),
				userStruct, n.element.Name)
			fmt.Fprintf(w, "%s\t\t}\n", tab)
		}

	}
	for _, child := range n.children {
		c := child.(*NodeData)
		processNodeForSetBulk(w, c, loopVar)
	}
	if addedLoop {
		fmt.Fprintf(w, "\t\t}\n")
	}
}

func checkIfLoopRequired(n *NodeData) bool {
	loopRequired := false
	for _, child := range n.children {
		c := child.(*NodeData)
		switch c.kind {
		case "leaf":
			return c.element.Value == ""
		case "leaf-list", "list":
			return false
		default:
			loopRequired = loopRequired || checkIfLoopRequired(c)
		}
	}
	return loopRequired
}

//--------------------- Functions related to Get-bulk operations ---------------------------
/*
func processBulkGetStruct(ws io.Writer, fileName string, bs *BulkStructure) {
	processBulkStruct(ws, fileName, bs, "Filter")
	processBulkStruct(ws, fileName, bs, "Get")
}

func processBulkGetApi(w io.Writer, fileName string, bs *BulkStructure) {
	fmt.Fprintf(w, "\nfunc ProcessSetBulk%s(sID int, v *SetBulk%s) error {\n", bs.Name, bs.Name)
	fmt.Fprintf(w, "\n\tvar buffer, b []byte\n")
	fmt.Fprintf(w, "\tvar err error\n")
	processBulkMembers(w, fileName, bs, false)
	addGetRPCCall(w)
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}

func processBulkGetMembers(w io.Writer, fileName string, bs *BulkStructure) {
	members := strings.Split(bs.Structs, " ")
	for _, m := range members {
		s := getFromGetStructMap(m)
		membersInside := strings.Split(s.Members, " ")
		filterInside := strings.Split(s.Filter, " ")
		initNodesTree()
		for _, mIn := range membersInside {
			id, _ := strconv.Atoi(mIn)
			e := elementById[fileName][id]
			// fmt.Fprintf(ws, "\t%s %s\n", e.Name, e.Type)
			updateNodesTreeFromPath(e.Path, m, &e)
		}

		for _, nodeInfo := range RootNodes {
			fmt.Fprintf(w, "\t{\n")

			processNodeForSetBulk(w, nodeInfo, nil)

			fmt.Fprintf(w, "\t\tif len(v.%s) > 0 {\n", nodeInfo.structName)
			fmt.Fprintf(w, "\t\t\tb, err = nc.MarshalIndent(&%s, \"\", \"  \")\n", nodeInfo.varName)
			fmt.Fprintf(w, "\t\t\tif err != nil {\n")
			fmt.Fprintf(w, "\t\t\t\tlog.Printf(\"Failed to edit config: %%s\", err.Error())\n")
			fmt.Fprintf(w, "\t\t\t\treturn err\n")
			fmt.Fprintf(w, "\t\t\t}\n")
			fmt.Fprintf(w, "\t\t\tbuffer = append(buffer, b...)\n")
			fmt.Fprintf(w, "\t\t}\n")
			fmt.Fprintf(w, "\t}\n")
		}
		fmt.Fprintf(w, "\n\n")
	}
} */
