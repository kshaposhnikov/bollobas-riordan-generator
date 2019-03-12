package bollobasriordan

import (
	"github.com/kshaposhnikov/bollobas-riordan-generator/graph"
	"log"
	"math"
	"math/rand"
	"sort"

	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/floats"
)

type BRGenerator struct {
	VCount   int
	ECount   int
	coolRank []int
}

func NewBRGenerator(vCount int, eCount int) *BRGenerator {
	generator := BRGenerator{
		VCount:   vCount,
		ECount:   eCount,
		coolRank: make([]int, vCount*eCount),
	}

	for i := 0; i < len(generator.coolRank); i++ {
		generator.coolRank[i] = 100
	}

	return &generator
}

//bollobas-riordan
// Number of threads should be less then m
func (gen *BRGenerator) Generate() *graph.Graph {
	if gen.ECount < 2 {
		log.Fatal("ECount should more or equal 2")
	}

	var previousGraph = gen.buildInitialGraph()
	return gen.buildFinalGraph(previousGraph, 0, previousGraph.GetNodeCount(), int64(gen.ECount))
}

func (gen *BRGenerator) buildInitialGraph() *graph.Graph {
	previousGraph := graph.NewGraph()
	previousGraph.AddNode(graph.Node{
		Id:                   1,
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int64{1},
	})

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	degree := make(map[int]int)
	degree[0] = 2
	for i := 1; i <= gen.VCount*gen.ECount-1; i++ {
		previousGraph = gen.nextGraph(previousGraph, degree, random)
		if i%100 == 0 {
			logrus.Info("Iter i = ", i)
		}
		logrus.Debug("[simplegenerator.buildInitialGraph] Graph for n = ", i, *previousGraph)
	}

	return previousGraph
}

func (gen *BRGenerator) nextGraph(previousGraph *graph.Graph, degrees map[int]int, random *rand.Rand) *graph.Graph {
	probabilities := gen.mtCalculateProbabilities(degrees)
	cdf := gen.cumsum(probabilities)

	x := random.Float64()
	idx := sort.Search(len(cdf), func(i int) bool {
		return cdf[i] > x
	})

	degrees[idx]++
	gen.coolRank[idx]--

	degrees[len(probabilities)-1]++
	return previousGraph.AddNode(graph.Node{
		Id:                   int64(len(probabilities)),
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int64{int64(idx) + 1},
	})
}

func (gen *BRGenerator) buildFinalGraph(pregeneratedGraph *graph.Graph, from, to int, m int64) *graph.Graph {
	result := graph.NewGraph()

	left := int64(from)
	j := left/m + 1
	var right = j*m - 1
	var loops []int64
	var l int64 = 0
	for _, node := range pregeneratedGraph.Nodes[from:to] {
		for _, associatedVertex := range node.AssociatedNodes {
			if associatedVertex < right && associatedVertex > left {
				loops = append(loops, j)
			} else if associatedVertex >= right || associatedVertex <= left {
				result = result.AddAssociatedNodeTo(j, int64(gen.calculateInterval(int(associatedVertex), int(m))))
			}
		}

		if ((left+l+1)/m)+1 > j {
			if len(loops) > 0 {
				result = result.AddAssociatedNodesTo(j, loops)
			} else if !result.ContainsVertex(j) {
				result = result.AddNode(graph.Node{
					Id:                   j,
					AssociatedNodesCount: len(loops),
					AssociatedNodes:      loops,
				})
			}
			loops = []int64{}
			left = right + 1
			right += m
			j++
			l = -1
		}
		l++
	}

	return result
}

func (gen *BRGenerator) calculateInterval(number int, m int) int {
	if number%m == 0 {
		return number / m
	} else {
		return int(math.Trunc(float64(number)/float64(m)) + 1)
	}
}

const nodeRate = 10

func (gen *BRGenerator) mtCalculateProbabilities(degrees map[int]int) []float64 {
	if len(degrees) > runtime.NumCPU()*nodeRate {
		batch := gen.calculateInterval(len(degrees), runtime.NumCPU())
		goroutineNumber := gen.calculateInterval(len(degrees), batch)
		probabilityResults := make(chan probabilityResult, goroutineNumber)
		var wg sync.WaitGroup
		wg.Add(goroutineNumber)
		for i := 0; i < goroutineNumber; i++ {
			from := i * batch
			to := from + batch
			if to >= len(degrees) {
				to = len(degrees)
			}

			go func(order int) {
				defer wg.Done()
				probabilityResults <- probabilityResult{
					order,
					gen.calculateProbabilities(degrees, from, to),
				}
			}(i)
		}
		wg.Wait()
		close(probabilityResults)

		var probabilities []probabilityResult
		for result := range probabilityResults {
			probabilities = append(probabilities, result)
		}
		sort.Slice(probabilities, func(i, j int) bool {
			return probabilities[i].order > probabilities[j].order
		})
		var result []float64
		for _, item := range probabilities {
			result = append(result, item.probabilities...)
		}
		return result
	} else {
		return gen.calculateProbabilities(degrees, 0, len(degrees))
	}
}

func (gen *BRGenerator) calculateProbabilities(degrees map[int]int, from, to int) []float64 {
	n := float64(len(degrees) + 1)
	var probabilities []float64
	// Сделать кофэффициент большим для новой вершины и уменьшать по мере роста степени этой вершины
	for i := from; i < to; i++ {
		//probabilities = append(probabilities, float64(degrees[i])/(2.0*n-1.0))
		logrus.Info("I = ", i, " len(coolRank) = ", len(gen.coolRank))
		a := float64(gen.coolRank[i])
		probabilities = append(probabilities, (float64(degrees[i])-1+a)/((a+1.0)*n+1.0))
	}

	if to == len(degrees) {
		//	probabilities = append(probabilities, 1.0/(2.0*n-1.0))
		a := float64(gen.coolRank[len(degrees)])
		probabilities = append(probabilities, a/((a+1.0)*n+1.0))
	}

	return probabilities
}

func (gen *BRGenerator) cumsum(probabilities []float64) []float64 {
	dest := make([]float64, len(probabilities))
	return floats.CumSum(dest, probabilities)
}

type probabilityResult struct {
	order         int
	probabilities []float64
}
