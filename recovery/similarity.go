package recovery

import "sort"

type SimilarityGraph struct {
	*DependencyModel
	NodeSlice []*SimilarityNode
	Nodes     map[string]*SimilarityNode // Indexed by node name
}

type SimilarityNode struct {
	*DependencyNode
	SortedNeighbors SortedNeighbors            // Sorted by Similarity
	Neighbors       map[string]*SimilarityEdge // Key is the name of the destination node
}

type SimilarityEdge struct {
	Neighbor   *SimilarityNode
	Similarity float64
}

type SortedNeighbors []*SimilarityEdge

func (n SortedNeighbors) Len() int {
	return len(n)
}

func (n SortedNeighbors) Less(a, b int) bool {
	return n[a].Similarity < n[a].Similarity
}

func (n SortedNeighbors) Swap(a, b int) {
	n[a], n[b] = n[b], n[a]
}

func (m *DependencyModel) BuildSimilarityGraph(intraGroupSimilarity, intraLayerSimilarity float64) *SimilarityGraph {
	nodes := make(map[string]*SimilarityNode)
	nodeSlice := make([]*SimilarityNode, 0, len(m.Names))
	for name, depNode := range m.Names {
		node := &SimilarityNode{
			DependencyNode:  depNode,
			Neighbors:       make(map[string]*SimilarityEdge),
			SortedNeighbors: make(SortedNeighbors, 0, len(m.Names)),
		}
		nodes[name] = node
		nodeSlice = append(nodeSlice, node)
	}
	graph := &SimilarityGraph{
		DependencyModel: m,
		Nodes:           nodes,
		NodeSlice:       nodeSlice,
	}
	graph.setEdges(intraGroupSimilarity, intraLayerSimilarity)
	return graph
}

func (g *SimilarityGraph) setEdges(intraGroupSimilarity, intraLayerSimilarity float64) {
	nodes := g.NodeSlice
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			node1, node2 := nodes[i], nodes[j]
			similarity := node1.getSimilarity(node2, intraGroupSimilarity, intraLayerSimilarity)
			if similarity > 0 {
				node1.addNeighbor(similarity, node2)
				node2.addNeighbor(similarity, node1)
			}
		}
	}
	for _, node := range g.Nodes {
		sort.Sort(node.SortedNeighbors)
	}
}

func (node1 *SimilarityNode) getSimilarity(node2 *SimilarityNode, intraGroupSimilarity, intraLayerSimilarity float64) float64 {
	for _, group1 := range node1.Groups {
		for _, group2 := range node2.Groups {
			if group1 == group2 {
				// The nodes share a group -> use the largest similarity
				return intraGroupSimilarity
			}
		}
	}
	if node1.Layer == node2.Layer {
		// The nodes are on the same layer -> use second largest similarity
		return intraLayerSimilarity
	}
	return 0
}

func (node *SimilarityNode) addNeighbor(similarity float64, neighbor *SimilarityNode) {
	edge := &SimilarityEdge{
		Similarity: similarity,
		Neighbor:   neighbor,
	}
	node.Neighbors[neighbor.Name] = edge
	node.SortedNeighbors = append(node.SortedNeighbors, edge)
}
