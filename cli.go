package main

import (
	"github.com/deckarep/golang-set"
	"github.com/kshaposhnikov/bollobas-riordan-generator/generator"
	"github.com/kshaposhnikov/bollobas-riordan-generator/generator/bollobasriordan"
	"github.com/kshaposhnikov/bollobas-riordan-generator/graph"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2"
	"time"
)

var generateCmd = &cobra.Command{
	Use:              "generate-graph",
	Short:            "Generate graph using Bollobas-Riordan method",
	TraverseChildren: true,
}

var vCount int
var eCount int

var initialAttractiveness float64

var leftLimit float64
var rightLimit float64
var step float64

var dbName string
var collectionName string
var threadCount int
var samplesCount int
var isDebug bool

func init() {
	var brCommand = &cobra.Command{
		Use:   "bollobas-riordan",
		Short: "Generate graph using Bollobas-Riordan method",
		Run:   generateBollobasRiordan,
	}

	var boCommand = &cobra.Command{
		Use:   "buckley-osthus",
		Short: "Generate graph using Buckley-Osthus method",
		Run:   generateBuckleyOsthus,
	}

	var yamCommand = &cobra.Command{
		Use:   "modified-buckley-osthus",
		Short: "Generate graph using Modified Buckley-Osthus method",
		Run:   generateModifiedBuckleyOsthus,
	}

	generateCmd.AddCommand(brCommand, boCommand, yamCommand)

	brCommand.Flags().IntVar(&vCount, "vCount", 6, "Number of Vertexes (n)")
	brCommand.Flags().IntVar(&eCount, "eCount", 2, "Number of Edges (m)")

	boCommand.Flags().IntVar(&vCount, "vCount", 6, "Number of Vertexes (n)")
	boCommand.Flags().IntVar(&eCount, "eCount", 2, "Number of Edges (m)")
	boCommand.Flags().Float64Var(&initialAttractiveness, "InitialAttractiveness",
		0.47, "Initial Attractiveness int model Buckley-Osthus (a)")

	yamCommand.Flags().IntVar(&vCount, "vCount", 6, "Number of Vertexes (n)")
	yamCommand.Flags().IntVar(&eCount, "eCount", 2, "Number of Edges (m)")
	yamCommand.Flags().Float64Var(&leftLimit, "leftLimit", 1, "Left Limit for (a)")
	yamCommand.Flags().Float64Var(&rightLimit, "rightLimit", 0.47, "Right Limit for (a)")
	yamCommand.Flags().Float64Var(&step, "step", 0.01, "Step of (a) changes")

	generateCmd.Flags().StringVar(&dbName, "db", "", "MongoDB database name")
	generateCmd.Flags().StringVar(&collectionName, "storage", "bollobas_riordan", "MongoDB collection name")
	generateCmd.Flags().IntVar(&threadCount, "threads", 1, "Number of threads")
	generateCmd.Flags().IntVar(&samplesCount, "samplesCount", 10, "Samples count to generate")
	generateCmd.Flags().BoolVar(&isDebug, "debug", false, "Enable debug logs")
}

func generateBollobasRiordan(cmd *cobra.Command, args []string) {
	generate(bollobasriordan.NewBRMTGenerator(vCount, eCount, threadCount))
}

func generateBuckleyOsthus(cmd *cobra.Command, args []string) {
	generate(bollobasriordan.NewBOGenerator(vCount, eCount, initialAttractiveness, threadCount))
}

func generateModifiedBuckleyOsthus(cmd *cobra.Command, args []string) {
	generate(bollobasriordan.NewYAMGenerator(vCount, eCount, threadCount, leftLimit, rightLimit, step))
}

func generate(generator generator.Generator) {
	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{
		"n":       vCount,
		"m":       eCount,
		"threads": threadCount,
	}).Debug("[Generate Command] Input parameters")

	for i := 0; i < samplesCount; i++ {
		logrus.Info("Generate Sample #", i+1)
		start := time.Now()
		result := generator.Generate()

		logrus.WithField("duration", time.Now().Sub(start)).Info("[Generate Command] Generation is done")
		logrus.Debug("[Generate Command] Final Graph", result)

		if dbName != "" {
			storeToDatabase(result)
		}
	}

}

func removeSelfLoopAndMultipleEdges(graph *graph.Graph) {
	var emptyNodes []int

	for i := 0; i < len(graph.Nodes); i++ {
		node := &graph.Nodes[i]
		uniqueNodes := mapset.NewSet()
		for _, id := range node.AssociatedNodes {
			if node.Id != id {
				uniqueNodes.Add(id)
			}
		}

		if uniqueNodes.Cardinality() > 0 {
			var tmp []int
			for id := range uniqueNodes.Iter() {
				tmp = append(tmp, id.(int))
			}

			node.AssociatedNodes = tmp
			node.AssociatedNodesCount = len(tmp)
		} else {
			emptyNodes = append(emptyNodes, i)
		}
	}

	for _, i := range emptyNodes {
		graph.Nodes = append(graph.Nodes[:i], graph.Nodes[i+1:]...)
	}
}

func storeToDatabase(graph *graph.Graph) {
	session, _ := mgo.Dial("localhost")
	db := session.DB(dbName)
	defer session.Clone()

	e := db.C(collectionName).Insert(graph)

	if e != nil {
		logrus.Fatal("Error in the time of inserting graph", e)
	}
}
