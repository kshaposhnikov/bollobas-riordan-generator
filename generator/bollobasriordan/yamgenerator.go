package bollobasriordan

import (
	"github.com/kshaposhnikov/bollobas-riordan-generator/graph"
	"github.com/sirupsen/logrus"
	"math/rand"
	"sort"
)

type YAMGenerator struct {
	*BRMTGenerator
	LeftLimit             float64
	RightLimit            float64
	Step                  float64
	initialAttractiveness float64
}

func NewYAMGenerator(vCount, eCount, threadCount int, leftLimit, rightLimit, step float64) *YAMGenerator {
	return &YAMGenerator{
		BRMTGenerator:         NewBRMTGenerator(vCount, eCount, threadCount),
		LeftLimit:             leftLimit,
		RightLimit:            rightLimit,
		Step:                  step,
		initialAttractiveness: leftLimit,
	}
}

func (gen *YAMGenerator) nextGraph(previousGraph *graph.Graph, degrees map[int]int, random *rand.Rand) *graph.Graph {
	probabilities := gen.mtCalculateProbabilities(degrees)
	cdf := gen.cumsum(probabilities)

	x := random.Float64()
	idx := sort.Search(len(cdf), func(i int) bool {
		return cdf[i] > x
	})

	degrees[idx]++

	logrus.Debug("x: ", x, " idx: ", idx, " CDF: ", cdf, " probabilities: ", probabilities)

	tmp := gen.initialAttractiveness - gen.Step
	if tmp >= gen.RightLimit {
		gen.initialAttractiveness = tmp
	}

	logrus.Debug("len(probabilities)-1: ", len(probabilities)-1, " len(degrees): ", len(degrees))

	degrees[len(probabilities)-1]++
	return previousGraph.AddNode(graph.Node{
		Id:                   len(probabilities),
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int{idx + 1},
	})
}

func (gen *YAMGenerator) probabilitiesForOldVertex(vertexDegree int, n float64) float64 {
	return (float64(vertexDegree) + gen.initialAttractiveness - 1.0) / ((gen.initialAttractiveness+1.0)*n - 1.0)
}

func (gen *YAMGenerator) probabilitiesForLastVertex(n float64) float64 {
	return gen.initialAttractiveness / ((gen.initialAttractiveness+1.0)*n - 1.0)
}
