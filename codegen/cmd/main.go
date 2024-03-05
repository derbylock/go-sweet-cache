package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
)

func main() {
	err := process()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func process() error {
	fileSet := token.NewFileSet()

	fileName := "testdata/repository.txt"
	file, err := os.Open(fileName)
	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read file %s: %w", fileName, err)
	}
	f, err := parser.ParseFile(fileSet, fileName, bytes, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("can't create ast for file %s: %w", fileName, err)
	}
	fmt.Printf("Number of decls: %d", len(f.Decls))
	return nil
}
