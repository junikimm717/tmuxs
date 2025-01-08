package main

import (
	"flag"
	"fmt"
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
	depth := flag.Int("d", 2, "depth to search workspaces")
	flag.Parse()
	chosenDir := flag.Arg(0)
	if len(chosenDir) > 0 {
		launchFromArg(chosenDir)
		return
	}

	channel := make(chan string)
	workspaces := filepath.SplitList(os.Getenv("WORKSPACES"))
	go sendAllDirOptions(channel, workspaces, *depth)
	res := FuzzyPick(channel)

	err := OpenSession(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
