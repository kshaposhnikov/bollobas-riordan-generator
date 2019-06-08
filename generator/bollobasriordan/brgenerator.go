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
	VCount                int
	ECount                int
	coolRank              []float64
	initialAttractiveness float64
}

func NewBRGenerator(vCount int, eCount int) *BRGenerator {
	generator := BRGenerator{
		VCount:                vCount,
		ECount:                eCount,
		coolRank:              make([]float64, vCount*eCount),
		initialAttractiveness: 1,
	}

	for i := 0; i < len(generator.coolRank); i++ {
		generator.coolRank[i] = 0.47
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
	return gen.buildFinalGraph(previousGraph, 0, previousGraph.GetNodeCount(), gen.ECount)
}

func (gen *BRGenerator) buildInitialGraph() *graph.Graph {
	previousGraph := graph.NewGraph()
	previousGraph.AddNode(graph.Node{
		Id:                   1,
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int{1},
	})

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	degree := make(map[int]int)
	degree[0] = 2
	for i := 1; i < gen.VCount*gen.ECount; i++ {
		previousGraph = gen.nextGraph(previousGraph, degree, random)
		if i%100 == 0 {
			logrus.Debug("Iter i = ", i)
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

	logrus.Debug("x: ", x, " idx: ", idx, " CDF: ", cdf, " probabilities: ", probabilities)

	//tmp := gen.coolRank[idx] + 0.01
	tmp := gen.initialAttractiveness - 0.05
	if tmp >= 0.47 {
		//gen.coolRank[idx] = tmp
		gen.initialAttractiveness = tmp
	}

	//else {
	//	gen.coolRank[idx] = 1
	//}

	logrus.Debug("len(probabilities)-1: ", len(probabilities)-1, " len(degrees): ", len(degrees))

	degrees[len(probabilities)-1]++
	return previousGraph.AddNode(graph.Node{
		Id:                   len(probabilities),
		AssociatedNodesCount: 1,
		AssociatedNodes:      []int{idx + 1},
	})
}

func (gen *BRGenerator) buildFinalGraph(pregeneratedGraph *graph.Graph, from, to int, m int) *graph.Graph {
	result := graph.NewGraph()

	left := from
	j := left/m + 1
	var right = j*m - 1
	var loops []int
	var l int = 0
	for _, node := range pregeneratedGraph.Nodes[from:to] {
		for _, associatedVertex := range node.AssociatedNodes {
			if associatedVertex < right && associatedVertex > left {
				loops = append(loops, j)
			} else if associatedVertex >= right || associatedVertex <= left {
				result = result.AddAssociatedNodeTo(j, gen.calculateInterval(associatedVertex, m))
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
			loops = []int{}
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
		//a := float64(gen.coolRank[i])
		a := gen.initialAttractiveness
		logrus.Debug("a: ", a)
		//a := 0.47 // лучший коээфициент
		f := (float64(degrees[i]) + a - 1.0) / ((a+1.0)*n - 1.0)
		probabilities = append(probabilities, f)
	}

	if to == len(degrees) {
		//probabilities = append(probabilities, 1.0/(2.0*n-1.0))

		logrus.Debug("len(degerees): ", len(degrees))

		if len(gen.coolRank) <= len(degrees) {
			logrus.Fatal(len(gen.coolRank), len(degrees))
		}
		//a := float64(gen.coolRank[len(degrees)])
		a := gen.initialAttractiveness
		logrus.Debug("a to: ", a)
		f := a / ((a+1.0)*n - 1.0)
		probabilities = append(probabilities, f)
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
