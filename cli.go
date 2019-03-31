package main

import (
	"github.com/deckarep/golang-set"
	"github.com/kshaposhnikov/bollobas-riordan-generator/generator/bollobasriordan"
	"github.com/kshaposhnikov/bollobas-riordan-generator/graph"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var generateCmd = &cobra.Command{
	Use:   "generate-graph",
	Short: "Generate graph using Bollobas-Riordan method",
	Run:   generate,
}

var graphConfig string
var dbName string
var collectionName string
var threadCount int
var samplesCount int
var isDebug bool

func init() {
	generateCmd.Flags().StringVarP(&graphConfig, "config", "c", "", "Config has format \"NodeCount;EdgeCount\"")
	generateCmd.Flags().StringVarP(&dbName, "db", "d", "", "MongoDB database name")
	generateCmd.Flags().StringVarP(&collectionName, "storage", "s", "bollobas_riordan", "MongoDB collection name")
	generateCmd.Flags().IntVarP(&threadCount, "threads", "t", 1, "Number of threads")
	generateCmd.Flags().IntVarP(&samplesCount, "samplesCount", "", 10, "Samples count to generate")
	generateCmd.Flags().BoolVarP(&isDebug, "debug", "", false, "Enable debug logs")
}

func generate(cmd *cobra.Command, args []string) {
	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	template := regexp.MustCompile(`[0-9]+;[0-9]+`)
	if template.MatchString(graphConfig) {
		nm := strings.Split(template.FindString(graphConfig), ";")
		n, _ := strconv.Atoi(nm[0])
		m, _ := strconv.Atoi(nm[1])

		logrus.WithFields(logrus.Fields{
			"n":       n,
			"m":       m,
			"threads": threadCount,
		}).Debug("[Generate Command] Input parameters")

		for i := 0; i < samplesCount; i++ {
			logrus.Info("Generate Sample #", i+1)
			start := time.Now()
			result := bollobasriordan.NewBRMTGenerator(n, m, threadCount).Generate()

			//removeSelfLoopAndMultipleEdges(result)
			logrus.WithField("duration", time.Now().Sub(start)).Info("[Generate Command] Generation is done")
			logrus.Debug("[Generate Command] Final Graph", result)

			if dbName != "" {
				//	removeSelfLoopAndMultipleEdges(result)
				storeToDatabase(result)
			}
		}
	} else {
		logrus.Error("Need to specify format `n;m`")
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
