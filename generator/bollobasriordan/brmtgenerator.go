package bollobasriordan

import (
	"github.com/kshaposhnikov/bollobas-riordan-generator/graph"
	"github.com/sirupsen/logrus"
	"sync"
)

type BRMTGenerator struct {
	BRGenerator
	ThreadCount int
}

func NewBRMTGenerator(vCount int, eCount int, threadCount int) *BRMTGenerator {
	return &BRMTGenerator{
		BRGenerator: *NewBRGenerator(vCount, eCount),
		ThreadCount: threadCount,
	}
}

func (gen *BRMTGenerator) Generate() *graph.Graph {
	initialGraph := gen.BRGenerator.buildInitialGraph()
	logrus.Info("Initial building done")
	batch := gen.calculateInterval(gen.VCount*gen.ECount, gen.ThreadCount)
	goroutineNumber := gen.calculateInterval(initialGraph.GetNodeCount(), batch)
	graphs := make(chan *graph.Graph, goroutineNumber)
	var wg sync.WaitGroup
	wg.Add(goroutineNumber)
	for i := 0; i < goroutineNumber; i++ {
		from := i * batch
		to := from + batch
		if to >= initialGraph.GetNodeCount() {
			to = initialGraph.GetNodeCount()
		}
		go func() {
			defer wg.Done()
			graphs <- gen.BRGenerator.buildFinalGraph(initialGraph, from, to, int(gen.ECount))
		}()
	}
	wg.Wait()
	close(graphs)

	result := graph.NewGraph()
	for item := range graphs {
		result.Concat(item)
	}

	logrus.Info("Building Done")
	return result
}
