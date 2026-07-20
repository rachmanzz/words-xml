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
	examplesDir := filepath.Join(root, "examples")
	outputDir := filepath.Join(root, "examples", "outputs")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading examples dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	modes := []string{"semantic", "lossless"}
	total := 0
	passed := 0
	failed := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".docx") {
			continue
		}

		inputPath := filepath.Join(examplesDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ".docx")

		fmt.Printf("=== %s ===\n", entry.Name())

		for _, mode := range modes {
			total++
			result, err := words.ProcessDOCXFileMode(inputPath, mode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [%s] ERROR: %v\n", mode, err)
				failed++
				continue
			}

			vr := words.Verify(result.WordsXML)
			if !vr.Valid {
				fmt.Fprintf(os.Stderr, "  [%s] INVALID:\n", mode)
				for _, e := range vr.Errors {
					fmt.Fprintf(os.Stderr, "    - %s\n", e)
				}
				failed++
				continue
			}

			if len(vr.Warns) > 0 {
				fmt.Printf("  [%s] OK (warnings: %d)\n", mode, len(vr.Warns))
				for _, w := range vr.Warns {
					fmt.Printf("    - %s\n", w)
				}
			} else {
				fmt.Printf("  [%s] OK\n", mode)
			}

			outFile := filepath.Join(outputDir, fmt.Sprintf("%s.%s.xml", baseName, mode))
			if err := os.WriteFile(outFile, []byte(result.WordsXML), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  [%s] WRITE ERROR: %v\n", mode, err)
				failed++
				continue
			}
			fmt.Printf("  [%s] -> %s\n", mode, outFile)
			passed++
		}
		fmt.Println()
	}

	fmt.Printf("Result: %d/%d passed, %d failed\n", passed, total, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
