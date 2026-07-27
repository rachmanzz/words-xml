package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rachmanzz/words-xml/words"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
}

func main() {
	root := projectRoot()
	dir := filepath.Join(root, "examples", "outputs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	total, passed, failed := 0, 0, 0

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".xml") {
			continue
		}
		total++
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("  %s: read error: %v\n", e.Name(), err)
			failed++
			continue
		}

		result := words.Verify(string(data))
		if result.Valid && len(result.Warns) == 0 {
			fmt.Printf("  %s: OK\n", e.Name())
			passed++
		} else if result.Valid {
			fmt.Printf("  %s: OK (warnings: %d)\n", e.Name(), len(result.Warns))
			for _, w := range result.Warns {
				fmt.Printf("    - %s\n", w)
			}
			passed++
		} else {
			fmt.Printf("  %s: FAIL\n", e.Name())
			for _, e := range result.Errors {
				fmt.Printf("    - %s\n", e)
			}
			for _, w := range result.Warns {
				fmt.Printf("    - %s\n", w)
			}
			failed++
		}
	}

	fmt.Printf("\nResult: %d/%d passed, %d failed\n", passed, total, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
