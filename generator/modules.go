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

// Create the main file that includes all data that is instantiated by
// different modules as one structure that represents the device.
func openMainFile(outdir string) *os.File {
	outpath := outdir + "/" + package_name + "/main.go"
	w, err := os.OpenFile(outpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		errorlog("unable to open file %s", err.Error())
		return nil
	}
	return w
}

// The header for the main file. Over time this will fill out though it
// is a small one for now :)
func mainFileHeader(w io.Writer) {
	fmt.Fprintf(w, "package yang\n")
}

// The structure includes all data that must be instantiated for
// representing the device. The data is aggregated from all the modules
// that have been compiled together.
func writeStructure(w io.Writer) {
	fmt.Fprintf(w, "type Device struct {\n")
	for _, m := range(modulesByName) {
		for _, sm := range m.submodules {
			addSubmodule(w, sm)
		}
	}
	fmt.Fprintf(w, "}\n")
}

// We generate all data that is instantiated at the level of the
// submodule. The groupings are instantiated using "uses" statement
// while the others are instantiated by their presence at the level
// of the module/submodule.
func addSubmodule(w io.Writer, sm *SubModule) {
	ymod := sm.module
	for _, u := range ymod.Uses {
		name := genTN(ymod, u.NName())
		fmt.Fprintf(w, "\t%s\n", name)
	}
	for _, cont := range ymod.Container {
		name := genTN(ymod, cont.NName())
		fmt.Fprintf(w, "\t%s %s\n", name, name)
	}
	for _, list := range ymod.List {
		name := genTN(ymod, list.NName())
		fmt.Fprintf(w, "\t%s %s\n", name, name)
	}
	for _, leaf := range ymod.Leaf {
		name := genTN(ymod, leaf.NName())
		fmt.Fprintf(w, "\t%s %s\n", name, name)
	}
	for _, leaflist := range ymod.LeafList {
		name := genTN(ymod, leaflist.NName())
		fmt.Fprintf(w, "\t%s %s\n", name, name)
	}
}

