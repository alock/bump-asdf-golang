package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get working directory")
	}
	var workspaceFiles []string
	var toolversionFiles []string
	err = filepath.Walk(workingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.Name() == ".tool-versions" {
			toolversionFiles = append(toolversionFiles, path)
		}
		if info.Name() == "workspace.xml" && strings.Contains(path, ".idea") {
			workspaceFiles = append(workspaceFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalln("failed to walk the path finding the files")
	}
	fmt.Println("workspace files:")
	fmt.Println(strings.Join(workspaceFiles, ","))
	fmt.Println("tool-versions files:")
	fmt.Println(strings.Join(toolversionFiles, ","))
}
