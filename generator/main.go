package main
import (
	"log"
	"strings"
	"io/ioutil"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pborman/getopt"
)

// Read the files from the directory
func readDir(path string, suffix string) []string {
	var filelist []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic("Error:" + err.Error())
	}
	for _, file := range files {
		if file.IsDir() {
			x := readDir(path+"/"+file.Name(), suffix)
			filelist = append(filelist, x...)
		} else {
			name := file.Name()
			if strings.HasSuffix(name, suffix) {
				filelist = append(filelist, path+"/"+name)
			}
		}
	}
	return filelist
}

func addModules(modules *yang.Modules) {
	for _, m := range modules.Modules {
		addModule(m)
	}
	for _, m := range modules.SubModules {
		addSubModule(m)
	}
}

func printModules() {
	for _, mod := range modulesByName {
		printModule(mod)
	}
}

func main() {
	var indir, outdir, apiIndir string
	getopt.StringVarLong(&indir, "indir", 'i', "directory to look for yang files")
	getopt.StringVarLong(&outdir, "outdir", 'o', "directory for output files")
	getopt.StringVarLong(&apiIndir, "api-indir", 'I', "directory for input api files")
	getopt.Parse()

	if indir == "" {
		log.Fatalf("-i: input directory for yang files must be present")
	}
	if outdir == "" {
		log.Fatalf("-o: output directory must be present")
	}

	// We recursively go through the directory for all the yang files which will
	// be included in the generated. We look for files named ".yang". We parse
	// those files and the output of parsing is stored in structure Modules defined
	// in package "yang".
	files := readDir(indir, "yang")
	debuglog("Number files = %d", len(files))
	ms := yang.NewModules()
	for _, file := range files {
		err := ms.Read(file)
		if err != nil {
			errorlog("Cannot open file: %s", err.Error())
		}
	}
	// Add all the modules parsed
	addModules(ms)

	// We have two steps in the overall processing of the modules which will
	// translate the modules to code. The first step is preprocess which attempts
	// to process some identities (mostly augments) that are to be used during
	// generation.
	for _, m := range modulesByName {
		m.preprocessModule()
	}

	// This generates the structure that describes the device based on yang
	// files included in the generation.
	generateMain(outdir)


	// Now generate code for each module. We generate a .go file for each
	// yang module
	/*
	for _, m := range modulesByName {
		processModule(m, outdir)
	}
	*/
	m := modulesByName["openconfig-optical-amplifier"]
	processModule(m, outdir)


	/*
        if apiIndir != "" {
                processStructsAndApis(apiIndir, outdir)
        }
	*/
}
