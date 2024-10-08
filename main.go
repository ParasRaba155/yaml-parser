package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		msg := `usage: yaml-parser <filename>`
		fmt.Println(msg)
		os.Exit(1)
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("couldn't open the given file: %s\n", err.Error())
		os.Exit(1)
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("couldn't read the given file: %s\n", err)
		os.Exit(1)
	}

	parser := NewParser(fileContent)
	yaml, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v", yaml)
	os.Exit(0)
}
