package main

import (
	"os"
	"io"
	"fmt"
)

// This file implements the global structure that describes the device
// by putting together all the uses declared within each of the module
// which essentially instantiate the device related paramters
func generateMain(outdir string) {
	w := openMainFile(outdir)
	if w == nil {
		return
	}
	defer w.Close()

	mainFileHeader(w)
	writeStructure(w)
}

func openMainFile(outdir string) *os.File {
	outpath := outdir + "/" + modulename + "/main.go"
	w, err := os.OpenFile(outpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		errorlog("unable to open file %s", err.Error())
		return nil
	}
	return w
}

func mainFileHeader(w io.Writer) {
	fmt.Fprintf(w, "package yang\n")
}

func writeStructure(w io.Writer) {
	fmt.Fprintf(w, "type Device struct {\n")
	for _, m := range(modulesByName) {
		for _, sm := range m.submodules {
			addSubmodule(w, sm)
		}
	}
	fmt.Fprintf(w, "}\n")
}

func addSubmodule(w io.Writer, sm *SubModule) {
	ymod := sm.module
	for _, u := range ymod.Uses {
		name := genTN(ymod, u.NName())
		fmt.Fprintf(w, "\t%s\n", name)
	}
}

