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
	// WIP: Just pass the fileContent
	fmt.Printf(">>>>>>>>>>\n%s>>>>>>>>>>\n", fileContent)
}
