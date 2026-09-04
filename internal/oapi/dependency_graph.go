package oapi

import (
	"sort"
	"strings"
)

// OperationNode represents a single API operation in the dependency graph.
type OperationNode struct {
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	Produces []string `json:"produces"`
	Consumes []string `json:"consumes"`
}

// DependencyEdge represents a data-flow link between two operations.
type DependencyEdge struct {
	From       OperationNode `json:"from"`
	To         OperationNode `json:"to"`
	FieldName  string        `json:"fieldName"`
	Confidence float64       `json:"confidence"`
}

// BuildDependencyGraph analyzes API specs to find data-flow dependencies between operations.
// It returns edges sorted by confidence descending, and topologically ordered groups.
func BuildDependencyGraph(specs []*APISpec) ([]DependencyEdge, [][]OperationNode) {
	nodes := buildNodes(specs)
	edges := findEdges(nodes)
	groups := topoSort(nodes, edges)
	return edges, groups
}

func buildNodes(specs []*APISpec) []OperationNode {
	nodes := make([]OperationNode, 0, len(specs))
	for _, spec := range specs {
		node := OperationNode{
			Method: string(spec.Method),
			Path:   spec.Path,
		}
		for _, prop := range spec.Response.Body {
			node.Produces = append(node.Produces, collectFieldNames(prop, "")...)
		}
		for _, prop := range spec.Request.PathParams {
			node.Consumes = append(node.Consumes, normalizeName(prop.Name))
		}
		for _, prop := range spec.Request.QueryParams {
			node.Consumes = append(node.Consumes, normalizeName(prop.Name))
		}
		for _, prop := range spec.Request.Body {
			node.Consumes = append(node.Consumes, collectFieldNames(prop, "")...)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func collectFieldNames(prop Property, prefix string) []string {
	name := normalizeName(prop.Name)
	if prefix != "" {
		name = prefix + "." + name
	}
	names := []string{name}
	for _, child := range prop.Children {
		names = append(names, collectFieldNames(child, name)...)
	}
	return names
}

func findEdges(nodes []OperationNode) []DependencyEdge {
	var edges []DependencyEdge
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			from := &nodes[i]
			to := &nodes[j]
			for _, produced := range from.Produces {
				for _, consumed := range to.Consumes {
					confidence := matchConfidence(produced, consumed, from.Path, to.Path)
					if confidence > 0 {
						edges = append(edges, DependencyEdge{
							From:       *from,
							To:         *to,
							FieldName:  produced,
							Confidence: confidence,
						})
					}
				}
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Confidence > edges[j].Confidence
	})
	return edges
}

func matchConfidence(produced, consumed, fromPath, toPath string) float64 {
	pNorm := normalizeName(produced)
	cNorm := normalizeName(consumed)

	if pNorm == cNorm {
		return 1.0
	}

	if singularPlural(pNorm, cNorm) {
		return 0.9
	}

	if isIDField(pNorm) && pathConsumesID(toPath, fromPath) {
		return 0.8
	}

	if pNorm == "id" && strings.Contains(cNorm, "id") {
		return 0.7
	}

	return 0
}

func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	return name
}

func singularPlural(a, b string) bool {
	if a+"s" == b || b+"s" == a {
		return true
	}
	if strings.TrimSuffix(a, "es") == b || strings.TrimSuffix(b, "es") == a {
		return true
	}
	return false
}

func isIDField(name string) bool {
	return name == "id" || strings.HasSuffix(name, "id")
}

func pathConsumesID(toPath, fromPath string) bool {
	fromResource := extractResource(fromPath)
	toResource := extractResource(toPath)
	if fromResource == "" || toResource == "" {
		return false
	}
	return fromResource == toResource || singularPlural(fromResource, toResource)
}

func extractResource(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if !strings.HasPrefix(parts[i], "{") {
			return strings.ToLower(parts[i])
		}
	}
	return ""
}

func topoSort(nodes []OperationNode, edges []DependencyEdge) [][]OperationNode {
	type nodeKey struct{ method, path string }
	key := func(n OperationNode) nodeKey { return nodeKey{n.Method, n.Path} }

	inDegree := make(map[nodeKey]int)
	adj := make(map[nodeKey][]nodeKey)
	nodeMap := make(map[nodeKey]OperationNode)

	for _, n := range nodes {
		k := key(n)
		inDegree[k] = 0
		nodeMap[k] = n
	}

	seen := make(map[[2]nodeKey]bool)
	for _, e := range edges {
		fk := key(e.From)
		tk := key(e.To)
		pair := [2]nodeKey{fk, tk}
		if seen[pair] {
			continue
		}
		seen[pair] = true
		adj[fk] = append(adj[fk], tk)
		inDegree[tk]++
	}

	var queue []nodeKey
	for k, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, k)
		}
	}
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].method+queue[i].path < queue[j].method+queue[j].path
	})

	var groups [][]OperationNode
	visited := make(map[nodeKey]bool)
	for len(queue) > 0 {
		var level []OperationNode
		var nextQueue []nodeKey
		for _, k := range queue {
			level = append(level, nodeMap[k])
			visited[k] = true
			for _, neighbor := range adj[k] {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					nextQueue = append(nextQueue, neighbor)
				}
			}
		}
		sort.Slice(nextQueue, func(i, j int) bool {
			return nextQueue[i].method+nextQueue[i].path < nextQueue[j].method+nextQueue[j].path
		})
		groups = append(groups, level)
		queue = nextQueue
	}

	// Collect nodes in cycles (unreached by topological sort).
	var remaining []OperationNode
	for k, n := range nodeMap {
		if !visited[k] {
			remaining = append(remaining, n)
		}
	}
	if len(remaining) > 0 {
		sort.Slice(remaining, func(i, j int) bool {
			return remaining[i].Method+remaining[i].Path < remaining[j].Method+remaining[j].Path
		})
		groups = append(groups, remaining)
	}

	return groups
}

// ApplyDependencies sets Order and NextRequest on scenarios based on dependency edges.
// It mutates the input specs by assigning execution order from the topological sort.
func ApplyDependencies(specs []*APISpec, edges []DependencyEdge) {
	type nodeKey struct{ method, path string }
	adj := make(map[nodeKey][]nodeKey)

	seen := make(map[[2]nodeKey]bool)
	for _, e := range edges {
		fk := nodeKey{e.From.Method, e.From.Path}
		tk := nodeKey{e.To.Method, e.To.Path}
		pair := [2]nodeKey{fk, tk}
		if seen[pair] {
			continue
		}
		seen[pair] = true
		adj[fk] = append(adj[fk], tk)
	}

	specMap := make(map[nodeKey]*APISpec)
	for _, spec := range specs {
		k := nodeKey{string(spec.Method), spec.Path}
		specMap[k] = spec
	}

	_, groups := BuildDependencyGraph(specs)
	for order, group := range groups {
		for _, node := range group {
			k := nodeKey{node.Method, node.Path}
			if spec, ok := specMap[k]; ok {
				spec.DependencyOrder = order
			}
		}
	}
}
