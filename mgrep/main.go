package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/asylcreek/mgrep2/worker"
	"github.com/asylcreek/mgrep2/worklist"
)

var args struct {
	SearchTerm string `arg:"positional,required"`
	SearchDir  string `arg:"positional"`
}

func main() {
	arg.MustParse(&args)

	var workersWg sync.WaitGroup

	worklist := worklist.New(100)

	results := make(chan worker.Result, 100)

	numWorkers := 10

	workersWg.Add(1)

	go func() {
		defer workersWg.Done()

		discoverDirs(&worklist, args.SearchDir)

		worklist.Finalize(numWorkers)
	}()

	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)

		go func() {
			defer workersWg.Done()

			for {
				path := worklist.Next().Path

				if path != "" {
					workerResult := worker.FindInFile(path, args.SearchTerm)

					if workerResult != nil {
						for _, line := range workerResult.Inner {
							results <- line
						}
					}
				} else {
					return
				}
			}
		}()
	}

	blockWorkersWg := make(chan struct{})

	go func() {
		workersWg.Wait()

		close(blockWorkersWg)
	}()

	var displayWg sync.WaitGroup

	displayWg.Add(1)

	go func() {
		for {
			select {
			case result := <-results:
				fmt.Printf("%v[%v]:%v\n", result.Path, result.LineNumber, result.Line)
			case <-blockWorkersWg:
				if len(results) == 0 {
					displayWg.Done()

					return
				}

			}

		}
	}()

	displayWg.Wait()
}

func discoverDirs(w *worklist.Worklist, path string) {
	entries, err := os.ReadDir(path)

	if err != nil {
		fmt.Println("Error: ", err)

		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			nextPath := filepath.Join(path, entry.Name())

			discoverDirs(w, nextPath)
		} else {
			w.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}
