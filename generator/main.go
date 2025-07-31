package main
import (
	"log"
	"fmt"
	"strings"
	"io/ioutil"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pborman/getopt"
)

var modules = []string {
	"openconfig-extensions",
	"openconfig-types",
	"openconfig-platform-types",
	"ietf-yang-types",
	"openconfig-inet-types",
	"openconfig-transport-types",
	"openconfig-yang-types",
	"ietf-interfaces",
	"openconfig-packet-match-types",
	"openconfig-mpls-types",
	"ietf-inet-types",
	"openconfig-interfaces",
	"openconfig-icmpv6-types",
	"openconfig-icmpv4-types",
	"openconfig-defined-sets",
	"openconfig-segment-routing-types",
	"openconfig-aft",
	"iana-if-type",
	"openconfig-bfd",
	"openconfig-keychain-types",
	"openconfig-packet-match",
	"openconfig-isis-types",
	"openconfig-srte-policy",
	"openconfig-if-ethernet",
	"openconfig-bgp-types",
	"openconfig-local-routing",
	"openconfig-aft-types",
	"openconfig-keychain",
	"openconfig-evpn-types",
	"openconfig-network-instance-types",
	"openconfig-acl",
	"openconfig-igmp-types",
	"openconfig-pim-types",
	"openconfig-segment-routing",
	"openconfig-if-aggregate",
	"openconfig-vlan-types",
	"openconfig-mpls-ldp",
	"openconfig-mpls-rsvp",
	"openconfig-rib-bgp",
	"openconfig-network-instance-static",
	"openconfig-pcep",
	"openconfig-evpn",
	"openconfig-igmp",
	"openconfig-pim",
	"openconfig-isis",
	"openconfig-policy-forwarding",
	"openconfig-ospf",
	"openconfig-ospfv2",
	"openconfig-vlan",
	"openconfig-mpls",
	"openconfig-bgp",
	"openconfig-network-instance-l3",
	"openconfig-network-instance",
	"openconfig-alarm-types",
	"openconfig-system-logging",
	"openconfig-platform",
	"openconfig-aaa-types",
	"openconfig-license",
	"openconfig-messages",
	"openconfig-alarms",
	"openconfig-procmon",
	"openconfig-system-terminal",
	"openconfig-aaa",
	"openconfig-wifi-types",
	"openconfig-system",
	"openconfig-platform-port",
	"openconfig-if-ip",
	"openconfig-ap-manager",
	"openconfig-wifi-mac",
	"openconfig-wifi-phy",
	"openconfig-system-grpc",
	"openconfig-lldp-types",
	"openconfig-platform-transceiver",
	"openconfig-if-8021x",
	"openconfig-if-tunnel",
	"openconfig-if-poe",
	"openconfig-access-points",
	"openconfig-ptp-types",
	"openconfig-ospf-types",
	"openconfig-gnsi",
	"openconfig-openflow-types",
	"openconfig-telemetry-types",
	"openconfig-fw-link-monitoring",
	"openconfig-oam",
	"openconfig-cfm-types",
	"openconfig-lldp",
	"openconfig-transport-line-common",
	"openconfig-sampling",
	"openconfig-probes-types",
	"openconfig-terminal-device-property-types",
	"openconfig-grpc-types",
	"openconfig-rib-bgp-types",
	"openconfig-qos",
	"ietf-yang-metadata",
	"openconfig-macsec-types",
	"openconfig-spanning-tree-types",
	"openconfig-p4rt",
	"openconfig-local-routing-network-instance",
	"openconfig-ap-interfaces",
	"openconfig-ptp",
	"openconfig-aft-network-instance",
	"openconfig-aft-summary",
	"openconfig-ospfv3-area-interface",
	"openconfig-ospf-policy",
	"openconfig-gnsi-pathz",
	"openconfig-gnsi-authz",
	"openconfig-gnsi-credentialz",
	"openconfig-gnsi-acctz",
	"openconfig-gnsi-certz",
	"openconfig-openflow",
	"openconfig-telemetry",
	"openconfig-qos-types",
	"openconfig-if-ip-ext",
	"openconfig-if-rates",
	"openconfig-if-sdn-ext",
	"openconfig-if-ethernet-ext",
	"openconfig-relay-agent",
	"openconfig-fw-high-availability",
	"openconfig-lacp",
	"openconfig-gribi",
	"openconfig-pf-srte",
	"openconfig-network-instance-policy",
	"openconfig-programming-errors",
	"openconfig-oam-cfm",
	"openconfig-wavelength-router",
	"openconfig-channel-monitor",
	"openconfig-terminal-device",
	"openconfig-transport-line-connectivity",
	"openconfig-optical-attenuator",
	"openconfig-optical-amplifier",
	"openconfig-transport-line-protection",
	"openconfig-rsvp-sr-ext",
	"openconfig-sampling-sflow",
	"openconfig-probes",
	"openconfig-ethernet-segments",
	"openconfig-bgp-policy",
	"openconfig-terminal-device-properties",
	"openconfig-gnpsi-types",
	"openconfig-rib-bgp-ext",
	"openconfig-hashing",
	"openconfig-system-utilization",
	"openconfig-system-bootz",
	"openconfig-system-controlplane",
	"openconfig-metadata",
	"openconfig-codegen-extensions",
	"openconfig-mpls-sr",
	"openconfig-macsec",
	"openconfig-platform-fan",
	"openconfig-platform-psu",
	"openconfig-platform-storage",
	"openconfig-platform-controller-card",
	"openconfig-platform-healthz",
	"openconfig-platform-pipeline-counters",
	"openconfig-platform-software",
	"openconfig-platform-linecard",
	"openconfig-platform-ext",
	"openconfig-platform-fabric",
	"openconfig-platform-cpu",
	"openconfig-platform-integrated-circuit",
	"openconfig-flexalgo",
	"openconfig-isis-lsdb-types",
	"openconfig-isis-policy",
	"openconfig-spanning-tree",
	"openconfig-ate-intf",
	"openconfig-ate-flow", }

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
	for _, mod := range modules {
		m, ok := modulesByName[mod]
		if ok {
			fmt.Println("Preprocessing module", mod, "....")
			m.preprocessModule()
		} else {
			fmt.Println("Didn't find module")
		}
	}
	/*
	for _, m := range modulesByName {
		fmt.Println("Preprocessing module", m.name, "....")
		m.preprocessModule()
	}
	*/

	// This generates the structure that describes the device based on yang
	// files included in the generation.
	generateMain(outdir)


	// Now generate code for each module. We generate a .go file for each
	// yang module
	fmt.Println("******        Start of processing of modules        ********")
	/*
	for _, mod := range modules {
		fmt.Println("Preprocessing module", m.name, "....")
		m, ok := modulesByName[mod]
		if ok {
			m.preprocessModule()
		} else {
			fmt.Println("Didn't find module")
		}
	}
	for _, m := range modulesByName {
		processModule(m, outdir)
	}
	m := modulesByName["openconfig-local-routing-network-instance"]
	processModule(m, outdir)
	*/
	/*
        if apiIndir != "" {
                processStructsAndApis(apiIndir, outdir)
        }
	*/
}
