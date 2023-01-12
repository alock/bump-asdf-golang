package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	asdfGoRootComponent        = `<component name="GOROOT" url="file://$USER_HOME$/.asdf/installs/golang/`
	golangSpace                = "golang "
	toolversionsFilename       = ".tool-versions"
	jetbrainsWorkspaceFilename = "workspace.xml"
)

var (
	version   = flag.String("v", "", "golang version to update the files with")
	updateAll = flag.Bool("a", false, "do not maintain minor versions, force update all")
	debugMode = flag.Bool("debug", false, "debug logs to help")
)

type fileToUpdate struct {
	filePathAndName      string // maybe make a file object
	currentGolangVersion string
}

func main() {
	flag.Parse()
	if !semver.IsValid(fmt.Sprintf("v%v", *version)) {
		log.Fatal("version flag is required and must be a valid semver")
	}
	log.Printf("version: %v, updateAll: %v\n", *version, *updateAll)
	workspaceFiles, toolversionsFiles := findFilesToUpdate()

	passedMajorMinor := semver.MajorMinor(fmt.Sprintf("v%v", *version))
	if len(workspaceFiles) > 0 {
		var wsFilesToUpdate []fileToUpdate
		for _, wsFileName := range workspaceFiles {
			majorMinor := semver.MajorMinor(fmt.Sprintf("v%v", wsFileName.currentGolangVersion))
			if *updateAll || passedMajorMinor == majorMinor {
				if *version != wsFileName.currentGolangVersion {
					wsFilesToUpdate = append(wsFilesToUpdate, wsFileName)
				}
			}
		}
		fmt.Printf("%v files to update\n", len(wsFilesToUpdate))
		for _, fileToBump := range wsFilesToUpdate {
			fmt.Printf("%v: %v\n", fileToBump.filePathAndName, fileToBump.currentGolangVersion)
		}
	}

	//make smarter since it's almost the same process
	if len(toolversionsFiles) > 0 {
		var asdfFilesToUpdate []fileToUpdate
		for _, asdfFileName := range toolversionsFiles {
			majorMinor := semver.MajorMinor(fmt.Sprintf("v%v", asdfFileName.currentGolangVersion))
			if *updateAll || passedMajorMinor == majorMinor {
				if *version != asdfFileName.currentGolangVersion {
					asdfFilesToUpdate = append(asdfFilesToUpdate, asdfFileName)
				}
			}
		}
		fmt.Printf("%v files to update\n", len(asdfFilesToUpdate))
		for _, fileToBump := range asdfFilesToUpdate {
			fmt.Printf("%v: %v\n", fileToBump.filePathAndName, fileToBump.currentGolangVersion)
		}
	}
}

func findFilesToUpdate() (workspaceFiles, toolversionsFiles []fileToUpdate) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory\n%v", err)
	}
	gopath := os.Getenv("GOPATH")
	err = filepath.Walk(workingDirectory, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}

		// do not process some files
		if gopath != "" && path == gopath && f.IsDir() {
			if *debugMode {
				log.Printf("skipping $GOPATH=%v and contents\n", path)
			}
			return filepath.SkipDir
		}
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || f.Name() == "node_modules") {
			if *debugMode {
				log.Printf("skipping %v and contents\n", path)
			}
			return filepath.SkipDir
		}
		if f.Name() == toolversionsFilename {
			currentVersion := getCurrentVersion(path)
			toolversionsFiles = append(toolversionsFiles, fileToUpdate{
				filePathAndName:      path,
				currentGolangVersion: currentVersion,
			})
		}
		if f.Name() == jetbrainsWorkspaceFilename && strings.Contains(path, ".idea") {
			currentVersion := getCurrentVersion(path)
			workspaceFiles = append(workspaceFiles, fileToUpdate{
				filePathAndName:      path,
				currentGolangVersion: currentVersion,
			})
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk the path finding the files\n%v", err)
	}
	return
}

func getCurrentVersion(filePathAndName string) string {
	fileContent, err := os.Open(filePathAndName)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(fileContent)
	for scanner.Scan() {
		if strings.Contains(filePathAndName, jetbrainsWorkspaceFilename) {
			if strings.Contains(scanner.Text(), asdfGoRootComponent) {
				return strings.Split(scanner.Text(), "/")[6]
			}
		}
		if strings.Contains(filePathAndName, toolversionsFilename) {
			if strings.Contains(scanner.Text(), golangSpace) {
				return strings.Split(scanner.Text(), " ")[1]
			}
		}
	}
	return ""
}

func rewriteFile(file fileToUpdate) {
	fileContents, err := os.ReadFile(file.filePathAndName)
	if err != nil {
		log.Fatal(err)
	}
	var updatedFileContents string
	if strings.Contains(file.filePathAndName, jetbrainsWorkspaceFilename) {
		updatedFileContents = strings.Replace(string(fileContents),
			fmt.Sprintf("%s%s/go", asdfGoRootComponent, file.currentGolangVersion),
			fmt.Sprintf("%s%s/go", asdfGoRootComponent, *version), 1)
	}
	if strings.Contains(file.filePathAndName, toolversionsFilename) {
		updatedFileContents = strings.Replace(string(fileContents),
			fmt.Sprintf("%s%s", golangSpace, file.currentGolangVersion),
			fmt.Sprintf("%s%s", golangSpace, *version), 1)
	}
	os.WriteFile(file.filePathAndName, []byte(updatedFileContents), 0644)
}
