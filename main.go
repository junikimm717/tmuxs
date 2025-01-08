package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func launchFromArg(arg string) {
	chosenDir, err := filepath.Abs(arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	stat, err := os.Stat(chosenDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Fprintln(os.Stderr, "Selected file is not a directory")
		os.Exit(1)
	}
	err = OpenSession(chosenDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}

func main() {
	depth := flag.Int("d", 3, "depth to search workspaces")
	parallelDepth := flag.Int("p", 1, "depth at which to start concurrently searching workspaces")
	flag.Parse()
	chosenDir := flag.Arg(0)
	if len(chosenDir) > 0 {
		launchFromArg(chosenDir)
		return
	}

	if *depth == 0 {
		*depth = math.MaxInt
	}

	if !ValidateDepth(*depth, *parallelDepth) {
		fmt.Fprintln(os.Stderr, "specified depth must be zero or strictly less than the parallel depth!")
		os.Exit(1)
	}

	channel := make(chan string)
	workspaces := filepath.SplitList(os.Getenv("WORKSPACES"))
	go sendAllDirOptions(channel, workspaces, *depth, *parallelDepth)
	res := FuzzyPick(channel)

	err := OpenSession(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
