//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxLinesPerFile   = 400
	maxExportsPerFile = 10
)

func main() {
	violations := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.Contains(path, "_generated") || strings.Contains(path, "generated_") {
			return nil
		}

		lines, lineErr := countLines(path)
		if lineErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, lineErr)
			return nil
		}

		if lines > maxLinesPerFile {
			fmt.Fprintf(os.Stderr, "FAIL: %s has %d lines (max %d)\n",
				path, lines, maxLinesPerFile)
			violations++
		}

		exports, exportErr := countExports(path)
		if exportErr != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, exportErr)
			return nil
		}

		if exports > maxExportsPerFile {
			fmt.Fprintf(os.Stderr, "FAIL: %s has %d exports (max %d)\n",
				path, exports, maxExportsPerFile)
			violations++
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if violations > 0 {
		fmt.Fprintf(os.Stderr, "\n%d file limit violation(s) found\n", violations)
		os.Exit(1)
	}

	fmt.Println("All file limits OK")
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

func countExports(path string) (int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0, err
	}

	exports := 0
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				exports++
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						exports++
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							exports++
						}
					}
				}
			}
		}
	}

	return exports, nil
}
