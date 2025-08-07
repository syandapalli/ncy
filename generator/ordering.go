package main

import (
	//"bufio"
	"fmt"
	//"os"
	//"regexp"
)

// ParseImports parses "import <module-name>;" from a YANG file
/*
func ParseImports(filepath string) ([]string, error) {
	var imports []string
	re := regexp.MustCompile(`^\s*import\s+([\w\-]+)\s*;`)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	return imports, scanner.Err()
}
*/

// BuildGraph builds a directed graph based on imports
func BuildGraph(modules map[string]*Module) (map[string][]string, map[string]int, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	//files := make(map[string]string)

	/*
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yang" {
			module := filepath.Base(path[:len(path)-5]) // Strip ".yang"
			files[module] = path
			graph[module] = []string{}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	*/

	for name, _ := range modules {
		graph[name] = []string{}
		inDegree[name] = 0
	}
	for name, module := range modules {
		debuglog("BuildGraph(): add module %s to graph", name)
		imports := getImports(module)
		debuglog("BuildGraph(): number of imports is %d for %s", len(imports), name)
		for _, imp := range imports {
			// Only track imports that are also present as local modules
			graph[imp] = append(graph[imp], name) // imp -> module
			inDegree[name]++
		}
	}

	return graph, inDegree, nil
}

// TopologicalSort returns modules in order respecting import dependencies
func TopologicalSort(graph map[string][]string, inDegree map[string]int) ([]string, error) {
	var order []string
	queue := []string{}

	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(graph) {
		return nil, fmt.Errorf("cycle detected in graph")
	}

	return order, nil
}

/*
func main() {
	dir := "/home/sriky/work/yangs/yang" // Replace with your actual YANG file directory
	graph, inDegree, err := BuildGraph(dir)
	if err != nil {
		fmt.Println("Error building graph:", err)
		return
	}

	order, err := TopologicalSort(graph, inDegree)
	if err != nil {
		fmt.Println("Error in topological sort:", err)
		return
	}

	fmt.Println("YANG modules in import-respecting order:")
	for _, module := range order {
		fmt.Println(module)
	}
}
*/

