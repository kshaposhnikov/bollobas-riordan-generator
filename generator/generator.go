package generator

import "github.com/kshaposhnikov/bollobas-riordan-generator/graph"

type (
	Generator interface {
		Generate() graph.Graph
	}
)
