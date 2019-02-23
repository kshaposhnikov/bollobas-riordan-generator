package main

import "runtime"

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use: "bollobas-riordan-generator",
}

func init() {
	initLogger()
	rootCmd.AddCommand(generateCmd)
}

func initLogger() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
