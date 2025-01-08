package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	fzf "github.com/junegunn/fzf/src"
)

var (
	ProjectRootIndicators = StringMap([]string{
		".git",
		"package.json",
		"license.md",
		"go.mod",
		"cargo.toml",
		"pyproject.toml",
		"requirements.txt",
	})

	IgnoreDirs = StringMap([]string{
		"node_modules",
		"__pycache__",
		"venv",
		"vendor",
	})
)

func IsProjectRootFile(name string) bool {
	name = strings.ToLower(name)
	return ProjectRootIndicators[name]
}

func ShouldIgnoreDir(name string) bool {
	return len(name) > 0 && (IgnoreDirs[name] || name[0] == '.')
}

type Entry struct {
	Path  string
	Depth int
}

func sendDirEntries(channel chan string, f fs.FS, prefix string, root string, depth int) {
	queue := []Entry{
		{Path: root, Depth: 0},
	}
	ptr := 0
	for ptr < len(queue) {
		cur := queue[ptr]
		ptr++
		channel <- filepath.Join(prefix, cur.Path)

		if depth != 0 && cur.Depth >= depth {
			continue
		}

		query := cur.Path
		if len(cur.Path) == 0 {
			query = "."
		}
		dirEntries, err := fs.ReadDir(f, query)
		if err != nil {
			continue
		}

		projectRoot := false
		for _, entry := range dirEntries {
			if IsProjectRootFile(entry.Name()) {
				projectRoot = true
				break
			}
		}
		if projectRoot {
			continue
		}

		for _, entry := range dirEntries {
			if entry.IsDir() && !ShouldIgnoreDir(entry.Name()) {
				prefix := cur.Path + string(os.PathSeparator)
				if len(cur.Path) == 0 {
					prefix = ""
				}
				queue = append(queue, Entry{
					Path:  prefix + entry.Name(),
					Depth: cur.Depth + 1,
				})
			}
		}
	}
}

func sendAllDirOptions(channel chan string, paths []string, depth int) {
	pathsMap := map[string]bool{}
	deDuplicated := make([]string, 0, len(paths))
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Println(err)
			continue
		}
		info, _ := os.Stat(path)
		if !pathsMap[path] {
			pathsMap[path] = true
			if info.IsDir() {
				deDuplicated = append(deDuplicated, path)
			}
		}
	}
	wg := sync.WaitGroup{}
	for _, root := range deDuplicated {
		wg.Add(1)
		go func() {
			sendDirEntries(channel, os.DirFS(root), root, "", depth)
			wg.Done()
		}()
	}
	wg.Wait()
	close(channel)
}

func FuzzyPick(inputChan chan string) string {
	options, _ := fzf.ParseOptions(
		true,
		[]string{},
	)

	outputChan := make(chan string)

	options.Input = inputChan
	options.Output = outputChan
	go func() {
		fzf.Run(options)
		// it seems I need to force close the channel or else the program just hangs.
		close(outputChan)
	}()

	return <-outputChan
}
