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
	Path    string
	AbsPath string
	Depth   int
}

type SendEntriesParams struct {
	Fs              fs.FS
	Prefix          string
	Root            string
	DontIncludeRoot bool
	Depth           int
}

func sendDirEntries(channel chan string, params SendEntriesParams) []Entry {
	queue := []Entry{
		{Path: params.Root, Depth: 0, AbsPath: filepath.Join(params.Prefix, params.Root)},
	}
	ptr := 0
	for ptr < len(queue) {
		cur := queue[ptr]
		ptr++
		if cur.Depth != 0 || !params.DontIncludeRoot {
			channel <- cur.AbsPath
		}

		if cur.Depth >= params.Depth {
			continue
		}

		query := cur.Path
		if len(cur.Path) == 0 {
			query = "."
		}
		dirEntries, err := fs.ReadDir(params.Fs, query)
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
					Path:    prefix + entry.Name(),
					Depth:   cur.Depth + 1,
					AbsPath: filepath.Join(cur.AbsPath, entry.Name()),
				})
			}
		}
	}
	return queue
}

func ValidateDepth(depth int, parallelDepth int) bool {
	return 1 <= parallelDepth && parallelDepth <= depth
}

func sendAllDirOptions(channel chan string, paths []string, depth int, parallelDepth int) {
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
			wg2 := sync.WaitGroup{}
			dirFS := os.DirFS(root)

			entries := sendDirEntries(channel, SendEntriesParams{
				Fs:     dirFS,
				Root:   "",
				Prefix: root,
				Depth:  parallelDepth,
			})

			for _, entry := range entries {
				if entry.Depth != parallelDepth || depth == parallelDepth {
					continue
				}
				wg2.Add(1)
				go func() {
					sendDirEntries(channel, SendEntriesParams{
						Fs:     dirFS,
						Prefix: root,
						Root:   entry.Path,
						Depth:  depth - parallelDepth,
						DontIncludeRoot: true,
					})
					wg2.Done()
				}()
			}
			wg2.Wait()
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
