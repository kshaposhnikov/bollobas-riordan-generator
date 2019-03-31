package graph

type Node struct {
	Id                   int
	AssociatedNodesCount int
	AssociatedNodes      []int
}

type Graph struct {
	Nodes []Node
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: []Node{},
	}
}

func (graph *Graph) Concat(newGraph *Graph) *Graph {
	for _, node := range newGraph.Nodes {
		graph.AddAssociatedNodesTo(node.Id, node.AssociatedNodes)
	}
	return graph
}

func (graph *Graph) AddNode(node Node) *Graph {
	graph.Nodes = append(graph.Nodes, node)
	return graph
}

func (graph *Graph) ContainsVertex(id int) bool {
	for _, node := range graph.Nodes {
		if node.Id == id {
			return true
		}
	}

	return false
}

func (graph *Graph) AddAssociatedNodesTo(id int, associatedNodes []int) *Graph {
	for _, associatedNode := range associatedNodes {
		graph.AddAssociatedNodeTo(id, associatedNode)
	}
	return graph
}

func (graph *Graph) AddAssociatedNodeTo(id int, associatedNodeName int) *Graph {
	for i, node := range graph.Nodes {
		if node.Id == id {
			graph.Nodes[i].AssociatedNodesCount++
			graph.Nodes[i].AssociatedNodes = append(graph.Nodes[i].AssociatedNodes, associatedNodeName)
			return graph
		}
	}

	return graph.AddNode(Node{
		Id:                   id,
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int{associatedNodeName},
	})
}

func (graph *Graph) GetNodeCount() int {
	return len(graph.Nodes)
}
