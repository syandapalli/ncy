package main
import (
	"fmt"
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

	files := readDir(indir, "yang")
	fmt.Println("Number files = ", len(files))
	ms := yang.NewModules()
	for _, file := range files {
		err := ms.Read(file)
		if err != nil {
			panic("Cannot open file: " + err.Error())
		}
	}
	addModules(ms)
	printModules()
	/*
        for _, sm := range ms.SubModules {
                fmt.Println("Submodule:", sm.Kind(), sm.NName(), sm.BelongsTo.Prefix.Name)
        }
        for _, m := range modulesByName {
                m.preprocessModule()
        }
        for _, m := range modulesByName {
                processModule(m, outdir)
        }

        if apiIndir != "" {
                processStructsAndApis(apiIndir, outdir)
        }
	*/
}
