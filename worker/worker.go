package worker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Result struct {
	Line       string
	LineNumber int
	Path       string
}

type Results struct {
	Inner []Result
}

func NewResult(line string, lineNumber int, path string) Result {
	return Result{line, lineNumber, path}
}

func FindInFile(path string, find string) *Results {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("Error", err)

		return nil
	}

	results := Results{make([]Result, 0)}

	scanner := bufio.NewScanner(file)

	lineNum := 0

	for scanner.Scan() {
		lineNum++

		line := scanner.Text()

		if strings.Contains(line, find) {
			newResult := NewResult(strings.TrimSpace(line), lineNum, path)

			results.Inner = append(results.Inner, newResult)
		}
	}

	if len(results.Inner) == 0 {
		return nil
	}

	return &results
}
