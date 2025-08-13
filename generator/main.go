package main
import (
	"log"
	"fmt"
	"strings"
	"io/ioutil"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pborman/getopt"
)

var ocmodules = []string {
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
	"openconfig-policy-types", // randomly inserted here
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
	"openconfig-ate-flow",
}

var bbfponmodules = []string {
	"bbf-inet-types",
	"bbf-subscriber-types",
	"bbf-yang-types",
	"bbf-xpon-onu-types",
	"bbf-xpon-defects",
	"bbf-xpon-onu-authentication-types",
	"ietf-inet-types",
	"bbf-xpon-power-management",
	"bbf-frame-processing-types",
	"bbf-qos-policing-types",
	"bbf-xpongemtcont",
	"bbf-xpon",
	"bbf-xponani",
	"iana-hardware",
	"bbf-device-types",
	"bbf-xponvani",
	"bbf-qos-types",
	"bbf-xpon-onu-authentication-features",
	"ietf-yang-types",
	"bbf-dot1q-types",
	"bbf-node-types",
	"bbf-frame-processing",
	"bbf-l2-dhcpv4-relay-profile-common",
	"bbf-xpon-types",
	"bbf-hardware-types",
	"bbf-xponvani-onu-authentication-groupings",
	"ietf-hardware",
	"ietf-interfaces",
	"bbf-frame-editing",
	"bbf-frame-classification",
	"bbf-device",
	"bbf-hardware-types-xpon",
	"bbf-hardware-transceivers",
	"ietf-hardware-state",
	"bbf-qos-traffic-mngt",
	"bbf-qos-policies-state",
	"bbf-link-table",
	"bbf-xponvani-onu-authentication",
	"bbf-hardware",
	"iana-if-type",
	"bbf-xpon-burst-profiles",
	"bbf-xponvani-power-management",
	"bbf-xponani-power-management",
	"bbf-xpon-onu-state",
	"bbf-qos-classifiers",
	"me-inventory",
	"bbf-hardware-transceivers-xpon",
	"bbf-xpongemtcont-qos",
	"bbf-if-type",
	"bbf-vlan-sub-interface-profiles",
	"bbf-interface-usage",
	"bbf-qos-policies",
	"bbf-l2-dhcpv4-relay",
	"bbf-sub-interfaces",
	"bbf-xpon-if-type",
	"bbf-vlan-sub-interface-profile-fp",
	"bbf-vlan-sub-interface-profile-usage",
	"bbf-qos-policing",
	"bbf-qos-policies-sub-interface-rewrite",
	"bbf-sub-interface-tagging",
	"bbf-xpon-onu-authentication",
	"bbf-qos-policer-envelope-profiles",
	"bbf-qos-policing-state",
	"bbf-frame-processing-profiles",
	"bbf-qos-policies-sub-interfaces",
}

var package_name string
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
	getopt.StringVarLong(&package_name, "package_name", 'p', "golang package name")
	getopt.StringVarLong(&apiIndir, "api-indir", 'I', "directory for input api files")
	getopt.Parse()

	if indir == "" {
		log.Fatalf("-i: input directory for yang files must be present")
	}
	if outdir == "" {
		log.Fatalf("-o: output directory must be present")
	}
	if  package_name == "" {
		package_name = "goyang"
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
	graph, inDegree, err := BuildGraph(modulesByName)
	if err != nil {
		fmt.Println("Error generating graph:", err)
		return
	}
	order, err := TopologicalSort(graph, inDegree)
	if err != nil {
		fmt.Println("Error in sort:", err)
	}
	//for id, name := range order {
	//	fmt.Println(id, ":", name)
	//}

	// We have two steps in the overall processing of the modules which will
	// translate the modules to code. The first step is preprocess which attempts
	// to process some identities (mostly augments) that are to be used during
	// generation.
	for _, mod := range order {
		m, ok := modulesByName[mod]
		if ok {
			fmt.Println("Preprocessing module", mod, "....")
			m.preprocessModule()
		} else {
			fmt.Println("Didn't find module", mod)
		}
	}

	// This generates the structure that describes the device based on yang
	// files included in the generation.
	//generateMain(outdir)


	// Now generate code for each module. We generate a .go file for each
	// yang module
	fmt.Println("******        Start of processing of modules        ********")
	for _, m := range modulesByName {
		processModule(m, outdir)
	}
	//m := modulesByName["openconfig-policy-types"]
	//processModule(m, outdir)
	
        //if apiIndir != "" {
	//	processStructsAndApis(apiIndir, outdir)
	//}
}
