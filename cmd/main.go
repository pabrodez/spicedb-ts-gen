package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	schemaConverter "github.com/pabrodez/spicedb-ts-gen/converter"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run main.go <folder path> <schema file name>")
	}
	folderPath := os.Args[1]
	schemaFileName := os.Args[2]

	if strings.TrimSpace(folderPath) == "" {
		log.Fatal("Folder path is empty")
	}

	if strings.TrimSpace(schemaFileName) == "" {
		log.Fatal("Schema file name is empty")
	}

	fmt.Print(schemaConverter.GenerateDefinitionFromFS(os.DirFS(folderPath), schemaFileName))
}
